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
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/segmentio/kafka-go"

	"go.opentelemetry.io/auto/internal/test/trigger"
)

const (
	imgName       = "kafka"
	containerName = "kafka_server"
	addr          = "127.0.0.1"
	port          = "9092"
)

var topics = []string{"topic1", "topic2"}

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
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	if err := pullKafkaImage(ctx, cli); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	if err := runKafkaContainer(ctx, cli); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	defer func() {
		if err := cleanupContainer(cli); err != nil {
			slog.Error("Failed to clean container", "error", err)
		}
	}()

	if err := streamContainerLogs(ctx, cli); err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	broker := addr + ":" + port
	if err = initKafka(ctx, broker, topics); err != nil {
		return fmt.Errorf("failed to initialize Kafka: %w", err)
	}
	slog.Info("Successfully setup kafka", "address", broker)

	readChan := reader(ctx, []string{broker})
	if err := produceMessages(ctx, broker); err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}
	// Wait for the read of the messages we just wrote.
	<-readChan

	return nil
}

func produceMessages(ctx context.Context, address string) error {
	kafkaWriter := &kafka.Writer{
		Addr:            kafka.TCP(address),
		Balancer:        &kafka.LeastBytes{},
		Async:           true,
		RequiredAcks:    1,
		WriteBackoffMax: 1 * time.Millisecond,
		BatchTimeout:    1 * time.Millisecond,
	}
	defer kafkaWriter.Close()

	return kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("key1"),
			Value: []byte("value1"),
			Topic: "topic1",
			Headers: []kafka.Header{
				{Key: "header1", Value: []byte("value1")},
			},
		},
		kafka.Message{
			Key:   []byte("key2"),
			Value: []byte("value2"),
			Topic: "topic2",
		},
	)
}

func reader(ctx context.Context, brokers []string) <-chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		defer close(done)

		cfg := kafka.ReaderConfig{
			Brokers:          brokers,
			GroupID:          "some group id",
			Topic:            "topic1",
			ReadBatchTimeout: 1 * time.Millisecond,
		}

		if err := cfg.Validate(); err != nil {
			panic(err)
		}
		reader := kafka.NewReader(cfg)
		defer reader.Close()

		slog.Info("Consuming ...")

		const maxRetries = 10
		for i := range maxRetries {
			select {
			case <-ctx.Done():
				return
			default:
			}

			slog.Info(
				"Attempting read...",
				"attempt", i+1,
				"maxAttempts", maxRetries,
			)
			_, err := reader.ReadMessage(ctx)
			if err != nil {
				slog.Error("Failed to read message", "error", err)
				continue
			}
			done <- struct{}{}
		}
	}()
	return done
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

func runKafkaContainer(ctx context.Context, cli *client.Client) error {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Hostname: addr, // Hacky, but works.
		User:     "root",
		Image:    kafkaImage,
		Env: []string{
			"KAFKA_CFG_NODE_ID=0",
			"KAFKA_CFG_PROCESS_ROLES=controller,broker",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@" + addr + ":9093",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE=true",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(port + "/tcp"): struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(port + "/tcp"): []nat.PortBinding{
				{HostIP: addr, HostPort: port},
			},
		},
	}, nil, nil, containerName)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	slog.Info("Container started", "image", kafkaImage, "name", containerName)
	return nil
}

func streamContainerLogs(ctx context.Context, cli *client.Client) error {
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

func cleanupContainer(cli *client.Client) error {
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
