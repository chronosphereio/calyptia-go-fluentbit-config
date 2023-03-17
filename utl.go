package fluentbitconfig

import (
	"bytes"
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

	if u, err := strconv.Unquote(s); err == nil {
		return u
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
		}
	case fmt.Stringer:
		return stringFromAny(t.String())
	case json.Marshaler:
		if b, err := t.MarshalJSON(); err == nil {
			return stringFromAny(string(b))
		}
	case map[string]any:
		var buff bytes.Buffer
		enc := json.NewEncoder(&buff)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(t); err == nil {
			return buff.String()
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

		if t == "" {
			return `""`
		}

		return t
	}

	return stringFromAny(fmt.Sprintf("%v", v))
}

// isFloatInt reports whether a float is an integer number
// with no fractional part.
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

func intFromAny(v any) (int, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		if int64(int(v)) == v {
			return int(v), true
		}
	case float32:
		return int(v), true
	case float64:
		if float64(int(v)) == v {
			return int(v), true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}
