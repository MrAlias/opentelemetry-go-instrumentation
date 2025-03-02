// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package process

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// ID represents a process ID.
type ID int

// Validate returns if id represents a valid running process.
func (id ID) Validate() error {
	if id < 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	p, err := os.FindProcess(int(id))
	if err != nil {
		return fmt.Errorf("no process with ID %d found: %w", id, err)
	}

	err = p.Signal(syscall.Signal(0))
	if err != nil {
		return fmt.Errorf("no process with ID %d found running: %w", id, err)
	}
	return nil
}

// ExecPath returns the executable path of the process ID.
func (id ID) ExecPath() (string, error) {
	path := fmt.Sprintf("/proc/%d/exe", id)
	return os.Readlink(path)
}

// BuildInfo returns the Go build info of the process ID executable.
func (id ID) BuildInfo() (*buildinfo.BuildInfo, error) {
	path, err := id.ExecPath()
	if err != nil {
		return nil, err
	}
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bi.GoVersion = strings.ReplaceAll(bi.GoVersion, "go", "")
	// Trims GOEXPERIMENT version suffix if present.
	if idx := strings.Index(bi.GoVersion, " X:"); idx > 0 {
		bi.GoVersion = bi.GoVersion[:idx]
	}

	return bi, nil
}
