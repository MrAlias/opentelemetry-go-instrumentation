// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otelsdk

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

type eBPFEventKeyType struct{}

var eBPFEventKey eBPFEventKeyType

// contextWithSpan returns a copy of parent in which span is stored.
func contextWithSpan(parent context.Context, span ptrace.Span) context.Context {
	return context.WithValue(parent, eBPFEventKey, span)
}

// spanFromContext returns the Span within ctx if one exists.
func spanFromContext(ctx context.Context) ptrace.Span {
	val := ctx.Value(eBPFEventKey)
	if val == nil {
		return ptrace.NewSpan()
	}

	s, _ := val.(ptrace.Span)
	return s
}
