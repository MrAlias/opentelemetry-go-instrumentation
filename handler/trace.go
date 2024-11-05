// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Tracer handles trace telemetry from auto-instrumentation.
type Tracer interface {
	// Trace handles the ScopeSpans passed from auto-instrumentation.
	Trace(context.Context, ptrace.ScopeSpans) error
	// Shutdown stops the Tracer.
	Shutdown(context.Context) error
}
