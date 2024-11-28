package target

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// SystemSource reads system processes and detects state changes.
type SystemSource struct {
	logger   *slog.Logger
	interval time.Duration
	prev     map[int]Process
}

// NewSystemSource creates a new SystemSource.
func NewSystemSource(logger *slog.Logger, interval time.Duration) *SystemSource {
	return &SystemSource{logger: logger, interval: interval}
}

// Start begins sending process state changes on the returned channel.
func (ss *SystemSource) Start(ctx context.Context) <-chan []ProcessState {
	output := make(chan []ProcessState)

	go func() {
		defer close(output)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(ss.interval):
				current, err := ss.fetchProcesses(ctx)
				if err != nil {
					ss.logger.Error("failed to fetch processes", "error", err)
					continue
				}

				changes := ss.detectChanges(current)
				if len(changes) > 0 {
					output <- changes
				}
				ss.prev = current
			}
		}
	}()
	return output
}

// fetchProcesses simulates fetching all processes from the system.
func (ss *SystemSource) fetchProcesses(ctx context.Context) (map[int]Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	out := make(map[int]Process, len(processes))
	for _, p := range processes {
		exe, err := p.ExeWithContext(ctx)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if exe == "/" {
			continue
		}

		out[int(p.Pid)] = Process{PID: int(p.Pid), Exec: exe}
	}
	return out, nil
}

// detectChanges finds created or removed processes.
func (ss *SystemSource) detectChanges(current map[int]Process) []ProcessState {
	changes := []ProcessState{}
	seen := make(map[int]bool)

	for pid, proc := range current {
		if _, exists := ss.prev[pid]; !exists {
			changes = append(changes, ProcessState{
				State:   StateCreated,
				Process: proc,
			})
		}
		seen[pid] = true
	}

	for pid, proc := range ss.prev {
		if !seen[pid] {
			changes = append(changes, ProcessState{
				State:   StateRemoved,
				Process: proc,
			})
		}
	}

	return changes
}
