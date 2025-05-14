package fluentbitconfig

import (
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

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

	if err := writePlugins("FILTER", c.Pipeline.Filters); err != nil {
		return nil, err
	}

	if err := writePlugins("OUTPUT", c.Pipeline.Outputs); err != nil {
		return nil, err
	}

	if err := writePlugins("PARSER", c.Parsers); err != nil {
		return nil, err
	}

	return []byte(sb.String()), nil
}

// reSpaces matches more than one consecutive spaces.
// This is used to split keys from values from the classic config.
var reSpaces = regexp.MustCompile(`\s+`)

func spaceCut(s string) (before, after string, found bool) {
	parts := reSpaces.Split(s, 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}
