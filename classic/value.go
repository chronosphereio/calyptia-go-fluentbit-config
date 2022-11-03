package classic

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// anyFromString used to convert property values from the classic config
// to any type.
func anyFromString(s string) any {
	// not using strconv since the boolean parser is not very strict
	// and allows: 1, t, T, 0, f, F.
	if strings.EqualFold(s, "true") {
		return true
	}

	if strings.EqualFold(s, "false") {
		return false
	}

	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}

	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}

	return s
}

// stringFromAny -
// TODO: Handle more data types.
func stringFromAny(v any) string {
	switch t := v.(type) {
	case encoding.TextMarshaler:
		if b, err := t.MarshalText(); err == nil {
			return stringFromAny(string(b))
		} else {
			return stringFromAny(fmt.Sprintf("%v", t))
		}
	case fmt.Stringer:
		return stringFromAny(t.String())
	case json.Marshaler:
		if b, err := t.MarshalJSON(); err == nil {
			return stringFromAny(string(b))
		} else {
			return stringFromAny(fmt.Sprintf("%v", t))
		}
	case map[string]any:
		if b, err := json.Marshal(t); err == nil {
			return string(b)
		} else {
			return stringFromAny(fmt.Sprintf("%v", t))
		}
	case float32:
		if isFloatInt(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return fmtFloat(t)
	case float64:
		if isFloatInt(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return fmtFloat(t)
	case string:
		if strings.Contains(t, "\n") {
			return fmt.Sprintf("%q", t)
		}
		return t
	default:
		return stringFromAny(fmt.Sprintf("%v", t))
	}
}

func isFloatInt[F float32 | float64](f F) bool {
	switch t := any(f).(type) {
	case float32:
		return t == float32(int32(f))
	case float64:
		return t == float64(int64(f))
	}
	return false
}

func fmtFloat[F float32 | float64](f F) string {
	s := fmt.Sprintf("%f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
