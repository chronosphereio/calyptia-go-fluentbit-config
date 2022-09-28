package classic

import (
	"fmt"
	"strconv"
	"strings"
)

type Value struct {
	Kind      ValueKind
	AsInt64   *int64
	AsFloat64 *float64
	AsBool    *bool
	AsString  *string
}

type ValueKind string

const (
	ValueKindInt64   ValueKind = "int64"
	ValueKindFloat64 ValueKind = "float64"
	ValueKindBool    ValueKind = "bool"
	ValueKindString  ValueKind = "string"
)

func (v Value) Int64() int64 {
	if v.Kind == ValueKindInt64 && v.AsInt64 != nil {
		return *v.AsInt64
	}
	return 0
}

func (v Value) Float64() float64 {
	if v.Kind == ValueKindFloat64 && v.AsFloat64 != nil {
		return *v.AsFloat64
	}
	return 0.0
}

func (v Value) Bool() bool {
	if v.Kind == ValueKindBool && v.AsBool != nil {
		return *v.AsBool
	}
	return false
}

func (v Value) String() string {
	switch {
	case v.Kind == ValueKindInt64 && v.AsInt64 != nil:
		return fmt.Sprintf("%d", *v.AsInt64)
	case v.Kind == ValueKindFloat64 && v.AsFloat64 != nil:
		return fmt.Sprintf("%f", *v.AsFloat64)
	case v.Kind == ValueKindBool && v.AsBool != nil:
		return fmt.Sprintf("%v", *v.AsBool)
	case v.Kind == ValueKindString && v.AsString != nil:
		return *v.AsString
	}
	return ""
}

func valueFromString(s string) Value {
	// not using strconv since the boolean parser is not very strict
	// and allows: 1, t, T, 0, f, F.
	if strings.EqualFold(s, "true") {
		return Value{
			Kind:   ValueKindBool,
			AsBool: ptr(true),
		}
	}

	if strings.EqualFold(s, "false") {
		return Value{
			Kind:   ValueKindBool,
			AsBool: ptr(false),
		}
	}

	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return Value{
			Kind:    ValueKindInt64,
			AsInt64: ptr(v),
		}
	}

	// float parser should come after int parser
	// cause it could catch integers as floats too.
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return Value{
			Kind:      ValueKindFloat64,
			AsFloat64: ptr(v),
		}
	}

	return Value{
		Kind:     ValueKindString,
		AsString: &s,
	}
}

func ptr[T any](v T) *T {
	return &v
}
