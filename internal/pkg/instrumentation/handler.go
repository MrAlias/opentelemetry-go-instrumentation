// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package instrumentation

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

type traceHandler interface {
	Trace(context.Context, ptrace.ScopeSpans) error
	Shutdown(context.Context) error
}
