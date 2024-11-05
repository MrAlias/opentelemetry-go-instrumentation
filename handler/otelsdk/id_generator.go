// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otelsdk

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type idGenerator struct{}

func newIDGenerator() *idGenerator {
	return &idGenerator{}
}

func (e *idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	s := spanFromContext(ctx)
	if s.TraceID().IsEmpty() || s.SpanID().IsEmpty() {
		return trace.TraceID{}, trace.SpanID{}
	}

	return trace.TraceID(s.TraceID()), trace.SpanID(s.SpanID())
}

func (e *idGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	return trace.SpanID(spanFromContext(ctx).SpanID())
}
