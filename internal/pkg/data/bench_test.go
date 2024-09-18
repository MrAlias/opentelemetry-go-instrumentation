// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

var (
	traceID = [16]byte{0x1}

	spanID1 = [8]byte{0x1}
	spanID2 = [8]byte{0x2}
	spanID3 = [8]byte{0x3}

	now      = time.Now()
	nowPlus1 = now.Add(1 * time.Second)

	attrA = []*KeyValue{
		{
			Key: "user",
			Value: &AnyValue{
				Value: &AnyValue_StringValue{
					StringValue: "Alice",
				},
			},
		},
		{
			Key: "admin",
			Value: &AnyValue{
				Value: &AnyValue_BoolValue{
					BoolValue: true,
				},
			},
		},
	}

	spanA = &Span{
		TraceId:           traceID,
		SpanId:            spanID2,
		ParentSpanId:      spanID1,
		Flags:             1,
		Name:              "span-a",
		StartTimeUnixNano: uint64(now.UnixNano()),
		EndTimeUnixNano:   uint64(nowPlus1.UnixNano()),
		Attributes:        attrA,
		Status: &Status{
			Message: "test status",
			Code:    Status_STATUS_CODE_OK,
		},
	}

	spanB = &Span{}

	scopeSpans = &ScopeSpans{
		Scope: &InstrumentationScope{
			Name:    "TestTracer",
			Version: "v0.1.0",
		},
		SchemaUrl: "http://go.opentelemetry.io/test",
		Spans:     []*Span{spanA, spanB},
	}
)

func BenchmarkJSONMarshalUnmarshal(b *testing.B) {
	var out ScopeSpans

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var inBuf bytes.Buffer
		enc := json.NewEncoder(&inBuf)
		err := enc.Encode(scopeSpans)
		if err != nil {
			b.Fatal(err)
		}

		payload := inBuf.Bytes()

		dec := json.NewDecoder(bytes.NewReader(payload))
		err = dec.Decode(&out)
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = out
}
