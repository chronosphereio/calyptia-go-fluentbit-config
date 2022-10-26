package classic

import (
	"fmt"
	"strconv"
	"strings"
)

// anyFromString can return bool, float64 and string.
func anyFromString(s string) any {
	// not using strconv since the boolean parser is not very strict
	// and allows: 1, t, T, 0, f, F.
	if strings.EqualFold(s, "true") {
		return true
	}

	if strings.EqualFold(s, "false") {
		return false
	}

	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}

	return s
}

// stringFromAny accepts bool, float64 and string.
func stringFromAny(v any) string {
	switch t := v.(type) {
	case float64:
		if isInteger(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return fmtFloat64(t)
	case string:
		if strings.Contains(t, "\n") {
			return fmt.Sprintf("%q", t)
		}
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func isInteger(f float64) bool {
	return f == float64(int64(f))
}

func fmtFloat64(f float64) string {
	s := fmt.Sprintf("%f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
