// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package kafka-go is a testing application for the
// [github.com/segmentio/kafka-go] package.
package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/segmentio/kafka-go"

	"go.opentelemetry.io/auto/internal/test/trigger"
)

const (
	imgName = "kafka"
	port    = "9092"
)

const (
	topic1 = "topic1"
	topic2 = "topic2"
)

var topics = []string{topic1, topic2}

//go:embed dependencies.Dockerfile
var dockerfile embed.FS

var kafkaImage = func() string {
	file, err := dockerfile.Open("dependencies.Dockerfile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 && fields[3] == imgName {
			return fields[1]
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	panic("image not found")
}()

func main() {
	var trig trigger.Flag
	flag.Var(&trig, "trigger", trig.Docs())
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Wait for auto-instrumentation.
	if err := trig.Wait(ctx); err != nil {
		slog.Error("Trigger failed", "error", err)
		os.Exit(1)
	}

	if err := run(ctx); err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	broker, clean, err := setup(ctx)
	if clean != nil {
		defer func() {
			if err := clean(); err != nil {
				slog.Error("Failed to clean up", "error", err)
			}
		}()
	}
	if err != nil {
		return fmt.Errorf("failed to setup: %w", err)
	}

	// Start readers before writing.
	readChan := reader(ctx, []string{broker})

	kafkaWriter := &kafka.Writer{
		Addr:            kafka.TCP(broker),
		Balancer:        &kafka.LeastBytes{},
		Async:           true,
		RequiredAcks:    1,
		WriteBackoffMax: 1 * time.Millisecond,
		BatchTimeout:    1 * time.Millisecond,
	}
	defer kafkaWriter.Close()

	msgs := []kafka.Message{
		{
			Key:   []byte("key1"),
			Value: []byte("value1"),
			Topic: topic1,
			Headers: []kafka.Header{
				{Key: "header1", Value: []byte("value1")},
			},
		},
		{
			Key:   []byte("key2"),
			Value: []byte("value2"),
			Topic: topic2,
		},
	}
	if err = kafkaWriter.WriteMessages(ctx, msgs...); err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	// Wait for the read of the messages we just wrote.
	for len(msgs) > 0 {
		var msg kafka.Message
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg = <-readChan:
		}
		slog.Debug(
			"Read message",
			"topic", msg.Topic,
			"key", string(msg.Key),
			"value", string(msg.Value),
			"headers", msg.Headers,
		)
		for i, m := range msgs {
			if m.Topic == msg.Topic && string(m.Key) == string(msg.Key) &&
				string(m.Value) == string(msg.Value) {
				msgs = slices.Delete(msgs, i, i+1)
				break
			}
		}
	}

	return nil
}

func reader(ctx context.Context, brokers []string) <-chan kafka.Message {
	out := make(chan kafka.Message)
	var wg sync.WaitGroup
	for _, topic := range topics {
		o := read(ctx, kafka.ReaderConfig{
			Brokers:          brokers,
			GroupID:          "some group id",
			Topic:            topic,
			ReadBatchTimeout: 1 * time.Millisecond,
		})

		wg.Add(1)
		go func(c <-chan kafka.Message) {
			defer wg.Done()
			for val := range c {
				out <- val
			}
		}(o)
	}

	// Close output channel once all input channels are drained
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func read(ctx context.Context, cfg kafka.ReaderConfig) <-chan kafka.Message {
	out := make(chan kafka.Message)
	go func() {
		defer close(out)

		reader := kafka.NewReader(cfg)
		defer reader.Close()

		for {
			slog.Info(
				"Consuming ...",
				"topic", cfg.Topic,
				"brokers", cfg.Brokers,
				"group", cfg.GroupID,
			)

			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				slog.Error("Failed to read message (skipping)", "error", err)
				continue
			}
			out <- msg
		}
	}()
	return out
}

func setup(ctx context.Context) (string, func() error, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	const containerName = "kafka"

	if err := pullKafkaImage(ctx, cli); err != nil {
		return "", nil, fmt.Errorf("failed to pull image: %w", err)
	}

	broker, err := runKafkaContainer(ctx, cli, containerName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to run container: %w", err)
	}

	cleanup := func() error { return cleanupContainer(cli, containerName) }

	if err := streamContainerLogs(ctx, cli, containerName); err != nil {
		return "", cleanup, fmt.Errorf("failed to stream logs: %w", err)
	}

	if err = initKafka(ctx, broker, topics); err != nil {
		return "", cleanup, fmt.Errorf("failed to initialize Kafka: %w", err)
	}
	slog.Info("Successfully setup kafka", "address", broker)

	return broker, cleanup, nil
}

func pullKafkaImage(ctx context.Context, cli *client.Client) (err error) {
	out, err := cli.ImagePull(ctx, kafkaImage, image.PullOptions{})
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err != nil {
			err = e
		}
	}()
	_, err = io.Copy(os.Stdout, out)
	return err
}

func runKafkaContainer(
	ctx context.Context,
	cli *client.Client,
	containerName string,
) (broker string, err error) {
	hostname := containerName
	broker = containerName + ":" + port
	hc := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(port + "/tcp"): []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	if n := getNetwork(ctx, cli); n != "" {
		slog.Info("Using parent container's network", "network", n)
		hc.NetworkMode = n
		// This is a hack to get the container to use the same network as the host.

		if n == "host" {
			hostname = "127.0.0.1"
			broker = "127.0.0.1:" + port
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Hostname: hostname,
		User:     "root",
		Image:    kafkaImage,
		Env: []string{
			"KAFKA_CFG_NODE_ID=0",
			"KAFKA_CFG_PROCESS_ROLES=controller,broker",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@" + hostname + ":9093",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE=true",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(port + "/tcp"): struct{}{},
		},
	}, hc, nil, nil, containerName)
	if err != nil {
		return broker, fmt.Errorf("failed to create container: %w", err)
	}
	slog.Info("Container created", "image", kafkaImage, "name", containerName)

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return broker, err
	}
	slog.Info("Container started", "image", kafkaImage, "name", containerName)
	return broker, nil
}

func isRunningInContainer() bool {
	_, err := os.ReadFile("/.dockerenv")
	if err == nil {
		fmt.Println("Running inside a Docker container.")
		return true
	}
	// fallback check using cgroup
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	fmt.Println("Cgroup data:", string(data))
	return strings.Contains(string(data), "docker")
}

func getContainerID() string {
	// fallback: hostname is usually container ID
	data, _ := os.ReadFile("/etc/hostname")
	hostname := strings.TrimSpace(string(data))

	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return hostname
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex to match the path pattern
	dockerPath := regexp.MustCompile(`/docker/containers/`)
	stripPrefix := regexp.MustCompile(`.*/docker/containers/`)
	stripSuffix := regexp.MustCompile(`/.*`)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		path := fields[3]
		if dockerPath.MatchString(path) {
			id := stripPrefix.ReplaceAllString(path, "")
			id = stripSuffix.ReplaceAllString(id, "")
			return id
		}
	}

	return hostname
}

func getNetwork(ctx context.Context, cli *client.Client) container.NetworkMode {
	if !isRunningInContainer() {
		return ""
	}

	containerJSON, err := cli.ContainerInspect(ctx, getContainerID())
	if err != nil {
		return ""
	}
	for name := range containerJSON.NetworkSettings.Networks {
		return container.NetworkMode(name)
	}
	return ""
}

func streamContainerLogs(ctx context.Context, cli *client.Client, containerName string) error {
	out, err := cli.ContainerLogs(
		ctx,
		containerName,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	go func() {
		defer out.Close()
		if _, err := io.Copy(os.Stdout, out); err != nil {
			slog.Error("Error streaming logs", "error", err)
		}
	}()

	return nil
}

func initKafka(ctx context.Context, address string, topics []string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	const maxRetries = 10
	for i := range maxRetries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			slog.Info(
				"Checking if Kafka is ready...",
				"attempt", i+1,
				"maxAttempts", maxRetries,
			)
		}

		var ready int
		for _, topic := range topics {
			conn, err := kafka.DialLeader(ctx, "tcp", address, topic, 0)
			if err != nil {
				slog.Info(
					"Kafka not ready",
					"error", err,
					"address", address,
					"topic", topic,
				)
			} else {
				ready++
				_ = conn.Close()
			}
		}

		if ready == len(topics) {
			return nil
		}
	}
	return errors.New("failed to initialize Kafka")
}

func cleanupContainer(cli *client.Client, containerName string) error {
	// Use our own context as the parent is likely already canceled.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	slog.Info("Stopping and removing container...", "name", containerName)
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop Kafka container: %w", err)
	}
	err := cli.ContainerRemove(
		ctx,
		containerName,
		container.RemoveOptions{Force: true},
	)
	if err != nil {
		return fmt.Errorf("failed to remove Kafka container: %w", err)
	}
	slog.Info("Container cleaned up", "name", containerName)
	return nil
}
