// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

var (
	ErrInvalidLengthCommon        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCommon          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCommon = fmt.Errorf("proto: unexpected end of group")
)

// AnyValue is used to represent any type of attribute value. AnyValue may contain a
// primitive value such as a string or integer or it may contain an arbitrary nested
// object containing arrays, key-value lists and primitives.
type AnyValue struct {
	// The value is one of the listed fields. It is valid for all values to be unspecified
	// in which case this AnyValue is considered to be "empty".
	//
	// Types that are assignable to Value:
	//	*AnyValue_StringValue
	//	*AnyValue_BoolValue
	//	*AnyValue_IntValue
	//	*AnyValue_DoubleValue
	//	*AnyValue_ArrayValue
	//	*AnyValue_KvlistValue
	//	*AnyValue_BytesValue
	Value isAnyValue_Value

	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AnyValue) UnmarshalJSON(data []byte) error {
	var raw struct{ Value map[string]any }
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	for _, val := range raw.Value {
		switch v := val.(type) {
		case string:
			m.Value = &AnyValue_StringValue{
				StringValue: v,
			}

		case bool:
			m.Value = &AnyValue_BoolValue{
				BoolValue: v,
			}
		case int64:
			m.Value = &AnyValue_IntValue{
				IntValue: v,
			}
		case float64:
			m.Value = &AnyValue_DoubleValue{
				DoubleValue: v,
			}
		case []byte:
			decoded, err := base64.StdEncoding.DecodeString(string(v))
			if err != nil {
				return fmt.Errorf("base64 decode: %w", err)
			}
			m.Value = &AnyValue_BytesValue{
				BytesValue: decoded,
			}
		/*
			 TODO:
						case "arrayValue", "array_value":
							m.Value = &AnyValue_ArrayValue{
								ArrayValue: readArray(iter),
							}
						case "kvlistValue", "kvlist_value":
							m.Value = &AnyValue_KvlistValue{
								KvlistValue: readKvlistValue(iter),
							}
		*/
		default:
			return fmt.Errorf("unsupported value type: %T", val)
		}
	}
	return nil
}

func (m *AnyValue) GetValue() isAnyValue_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *AnyValue) GetStringValue() string {
	if x, ok := x.GetValue().(*AnyValue_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (x *AnyValue) GetBoolValue() bool {
	if x, ok := x.GetValue().(*AnyValue_BoolValue); ok {
		return x.BoolValue
	}
	return false
}

func (x *AnyValue) GetIntValue() int64 {
	if x, ok := x.GetValue().(*AnyValue_IntValue); ok {
		return x.IntValue
	}
	return 0
}

func (x *AnyValue) GetDoubleValue() float64 {
	if x, ok := x.GetValue().(*AnyValue_DoubleValue); ok {
		return x.DoubleValue
	}
	return 0
}

func (x *AnyValue) GetArrayValue() *ArrayValue {
	if x, ok := x.GetValue().(*AnyValue_ArrayValue); ok {
		return x.ArrayValue
	}
	return nil
}

func (x *AnyValue) GetKvlistValue() *KeyValueList {
	if x, ok := x.GetValue().(*AnyValue_KvlistValue); ok {
		return x.KvlistValue
	}
	return nil
}

func (x *AnyValue) GetBytesValue() []byte {
	if x, ok := x.GetValue().(*AnyValue_BytesValue); ok {
		return x.BytesValue
	}
	return nil
}

type isAnyValue_Value interface {
	isAnyValue_Value()
}

type AnyValue_StringValue struct {
	StringValue string `json:"stringValue"`
}

type AnyValue_BoolValue struct {
	BoolValue bool `json:"boolValue"`
}

type AnyValue_IntValue struct {
	IntValue int64 `json:"intValue"`
}

type AnyValue_DoubleValue struct {
	DoubleValue float64 `json:"doubleValue"`
}

type AnyValue_ArrayValue struct {
	ArrayValue *ArrayValue `json:"arrayValue"`
}

type AnyValue_KvlistValue struct {
	KvlistValue *KeyValueList `json:"kvlistValue"`
}

type AnyValue_BytesValue struct {
	BytesValue []byte `json:"bytesValue"`
}

func (*AnyValue_StringValue) isAnyValue_Value() {}

func (*AnyValue_BoolValue) isAnyValue_Value() {}

func (*AnyValue_IntValue) isAnyValue_Value() {}

func (*AnyValue_DoubleValue) isAnyValue_Value() {}

func (*AnyValue_ArrayValue) isAnyValue_Value() {}

func (*AnyValue_KvlistValue) isAnyValue_Value() {}

func (*AnyValue_BytesValue) isAnyValue_Value() {}

// ArrayValue is a list of AnyValue messages. We need ArrayValue as a message
// since oneof in AnyValue does not allow repeated fields.
type ArrayValue struct {
	// Array of values. The array may be empty (contain 0 elements).
	Values []*AnyValue `json:"values,omitempty"`

	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (x *ArrayValue) GetValues() []*AnyValue {
	if x != nil {
		return x.Values
	}
	return nil
}

// KeyValueList is a list of KeyValue messages. We need KeyValueList as a message
// since `oneof` in AnyValue does not allow repeated fields. Everywhere else where we need
// a list of KeyValue messages (e.g. in Span) we use `repeated KeyValue` directly to
// avoid unnecessary extra wrapping (which slows down the protocol). The 2 approaches
// are semantically equivalent.
type KeyValueList struct {
	// A collection of key/value pairs of key-value pairs. The list may be empty (may
	// contain 0 elements).
	// The keys MUST be unique (it is not allowed to have more than one
	// value with the same key).
	Values []*KeyValue `json:"values,omitempty"`

	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (x *KeyValueList) GetValues() []*KeyValue {
	if x != nil {
		return x.Values
	}
	return nil
}

// KeyValue is a key-value pair that is used to store Span attributes, Link
// attributes, etc.
type KeyValue struct {
	Key   string    `json:"key,omitempty"`
	Value *AnyValue `json:"value,omitempty"`

	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (x *KeyValue) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KeyValue) GetValue() *AnyValue {
	if x != nil {
		return x.Value
	}
	return nil
}

// InstrumentationScope is a message representing the instrumentation scope information
// such as the fully qualified name and version.
type InstrumentationScope struct {
	// An empty instrumentation scope name means the name is unknown.
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	// Additional attributes that describe the scope. [Optional].
	// Attribute keys MUST be unique (it is not allowed to have more than one
	// attribute with the same key).
	Attributes             []*KeyValue `json:"attributes,omitempty"`
	DroppedAttributesCount uint32      `json:"dropped_attributes_count,omitempty"`

	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (x *InstrumentationScope) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *InstrumentationScope) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *InstrumentationScope) GetAttributes() []*KeyValue {
	if x != nil {
		return x.Attributes
	}
	return nil
}

func (x *InstrumentationScope) GetDroppedAttributesCount() uint32 {
	if x != nil {
		return x.DroppedAttributesCount
	}
	return 0
}

func skipCommon(data []byte) (n int, err error) {
	l := len(data)
	index := 0
	depth := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCommon
			}
			if index >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if index >= l {
					return 0, io.ErrUnexpectedEOF
				}
				index++
				if data[index-1] < 0x80 {
					break
				}
			}
		case 1:
			index += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if index >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCommon
			}
			index += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCommon
			}
			depth--
		case 5:
			index += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if index < 0 {
			return 0, ErrInvalidLengthCommon
		}
		if depth == 0 {
			return index, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}
