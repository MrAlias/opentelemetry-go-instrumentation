// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"
	"time"
)

var (
	traceID = [16]byte{0x1}

	spanID1 = [8]byte{0x1}
	spanID2 = [8]byte{0x2}

	now      = time.Now()
	nowPlus1 = now.Add(1 * time.Second)

	spanA = &Span{
		TraceId:           traceID,
		SpanId:            spanID2,
		ParentSpanId:      spanID1,
		Flags:             1,
		Name:              "span-a",
		StartTimeUnixNano: uint64(now.UnixNano()),
		EndTimeUnixNano:   uint64(nowPlus1.UnixNano()),
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

func BenchmarkGOBMarshalUnmarshal(b *testing.B) {
	// TODO:
	gob.Register(AnyValue_StringValue{})
	gob.Register(AnyValue_BoolValue{})

	var out ScopeSpans

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var inBuf bytes.Buffer
		enc := gob.NewEncoder(&inBuf)
		err := enc.Encode(scopeSpans)
		if err != nil {
			b.Fatal(err)
		}

		payload := inBuf.Bytes()

		dec := gob.NewDecoder(bytes.NewReader(payload))
		err = dec.Decode(&out)
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = out
}

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
