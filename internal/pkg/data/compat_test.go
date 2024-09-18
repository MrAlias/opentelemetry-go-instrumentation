// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var full = []byte(`{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {
              "stringValue": "my.service"
            }
          }
        ]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "my.library",
            "version": "1.0.0",
            "attributes": [
              {
                "key": "my.scope.attribute",
                "value": {
                  "stringValue": "some scope attribute"
                }
              }
            ]
          },
          "spans": [
            {
              "traceId": "5B8EFFF798038103D269B633813FC60C",
              "spanId": "EEE19B7EC3C1B174",
              "parentSpanId": "EEE19B7EC3C1B173",
              "name": "I'm a server span",
              "startTimeUnixNano": "1544712660000000000",
              "endTimeUnixNano": "1544712661000000000",
              "kind": 2,
              "attributes": [
                {
                  "key": "my.span.attr",
                  "value": {
                    "stringValue": "some value"
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}`)

var scopeSpan = []byte(`{
  "scope": {
    "name": "my.library",
    "version": "1.0.0",
    "attributes": [
      {
        "key": "my.scope.attribute",
        "value": {
          "stringValue": "some scope attribute"
        }
      }
    ]
  },
  "spans": [
    {
      "traceId": "5B8EFFF798038103D269B633813FC60C",
      "spanId": "EEE19B7EC3C1B174",
      "parentSpanId": "EEE19B7EC3C1B173",
      "name": "I'm a server span",
      "startTimeUnixNano": "1544712660000000000",
      "endTimeUnixNano": "1544712661000000000",
      "kind": 2,
      "attributes": [
        {
          "key": "my.span.attr",
          "value": {
            "stringValue": "some value"
          }
        }
      ]
    }
  ]
}`)

func TestJSONUnmarshal(t *testing.T) {
	dec := json.NewDecoder(bytes.NewReader(scopeSpan))

	var s ScopeSpans
	err := dec.Decode(&s)
	assert.NoError(t, err)
	t.Log(string(scopeSpan))
	t.Logf("scope: %#v", s.Scope)
	t.Logf("scope attribute: %#v", s.Scope.Attributes[0])

	t.Log("spans:")
	for _, span := range s.GetSpans() {
		t.Logf("%#v", span)
		t.Logf("attribute %#v", span.Attributes[0])
	}
}
