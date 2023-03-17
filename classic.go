package fluentbitconfig

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

// reSpaces matches more than one consecutive spaces.
// This is used to split keys from values from the classic config.
var reSpaces = regexp.MustCompile(`\s+`)

func (c *Config) UnmarshalClassic(text []byte) error {
	lines := strings.Split(string(text), "\n")

	type Section struct {
		Kind  SectionKind
		Props property.Properties
	}

	var currentSection *Section
	flushCurrentSection := func() {
		if currentSection == nil {
			return
		}

		c.AddSection(currentSection.Kind, currentSection.Props)

		currentSection = nil
	}

	for i, line := range lines {
		lineNumber := uint(i + 1)
		if !utf8.ValidString(line) {
			return NewLinedError("invalid utf8 string", lineNumber)
		}

		line = strings.TrimSpace(line)

		// ignore empty line
		if line == "" {
			continue
		}

		// ignore comment
		if strings.HasPrefix(line, "#") {
			continue
		}

		// command
		if strings.HasPrefix(line, "@") {
			flushCurrentSection()

			line := strings.TrimPrefix(line, "@")
			cmd, instruction, ok := spaceCut(line)
			if !ok {
				return NewLinedError("expected at least two strings separated by a space", lineNumber)
			}

			switch {
			case strings.EqualFold(cmd, "INCLUDE"):
				c.Include(instruction)
			case strings.EqualFold(cmd, "SET"):
				key, strValue, _ := strings.Cut(instruction, "=")
				c.SetEnv(key, anyFromString(strValue))
			}

			continue
		}

		// beginning of section
		if strings.HasPrefix(line, "[") {
			flushCurrentSection()

			if !strings.HasSuffix(line, "]") {
				return NewLinedError(`expected section to end with "]"`, lineNumber)
			}

			sectionName := strings.Trim(line, "[]")
			sectionName = strings.TrimSpace(sectionName)

			if sectionName == "" {
				return NewLinedError("expected section name to not be empty", lineNumber)
			}

			currentSection = &Section{
				Kind:  SectionKind(strings.ToLower(sectionName)),
				Props: property.Properties{},
			}

			continue
		}

		if currentSection == nil {
			return NewLinedError(fmt.Sprintf("unexpected entry %q", line), lineNumber)
		}

		key, strVal, ok := spaceCut(line)
		if !ok {
			return NewLinedError("expected at least two strings separated by a space", lineNumber)
		}

		value := anyFromString(strVal)

		// Case when the property could be repeated. Example:
		// 	[FILTER]
		// 		Name record_modifier
		// 		Match *
		// 		Record hostname ${HOSTNAME}
		// 		Record product Awesome_Tool
		if v, ok := currentSection.Props.Get(key); ok {
			if s, ok := v.([]any); ok {
				s = append(s, value)
				currentSection.Props.Set(key, s)
			} else {
				currentSection.Props.Set(key, []any{v, value})
			}
		} else {
			currentSection.Props.Add(key, value)
		}
	}

	flushCurrentSection()

	return nil
}

func (c Config) MarshalClassic() ([]byte, error) {
	var sb strings.Builder

	for _, p := range c.Env {
		_, err := fmt.Fprintf(&sb, "@SET %s=%s\n", p.Key, stringFromAny(p.Value))
		if err != nil {
			return nil, err
		}
	}

	for _, include := range c.Includes {
		_, err := fmt.Fprintf(&sb, "@INCLUDE %s\n", include)
		if err != nil {
			return nil, err
		}
	}

	writeProps := func(kind string, props property.Properties) error {
		if len(props) == 0 {
			return nil
		}

		_, err := fmt.Fprintf(&sb, "[%s]\n", kind)
		if err != nil {
			return err
		}

		tw := tabwriter.NewWriter(&sb, 0, 4, 1, ' ', 0)
		for _, p := range props {
			if s, ok := p.Value.([]any); ok {
				for _, v := range s {
					_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, stringFromAny(v))
					if err != nil {
						return err
					}
				}
			} else {
				_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, stringFromAny(p.Value))
				if err != nil {
					return err
				}
			}
		}
		return tw.Flush()
	}

	writePlugins := func(kind string, plugins Plugins) error {
		for _, plugin := range plugins {
			if err := writeProps(kind, plugin.Properties); err != nil {
				return err
			}
		}

		return nil
	}

	if err := writeProps("SERVICE", c.Service); err != nil {
		return nil, err
	}

	if err := writePlugins("CUSTOM", c.Customs); err != nil {
		return nil, err
	}

	if err := writePlugins("INPUT", c.Pipeline.Inputs); err != nil {
		return nil, err
	}

	if err := writePlugins("PARSER", c.Pipeline.Parsers); err != nil {
		return nil, err
	}

	if err := writePlugins("FILTER", c.Pipeline.Filters); err != nil {
		return nil, err
	}

	if err := writePlugins("OUTPUT", c.Pipeline.Outputs); err != nil {
		return nil, err
	}

	return []byte(sb.String()), nil
}

func spaceCut(s string) (before, after string, found bool) {
	parts := reSpaces.Split(s, 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

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
