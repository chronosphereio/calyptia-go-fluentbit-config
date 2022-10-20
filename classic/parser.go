package classic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	fluentbitconf "github.com/calyptia/go-fluentbit-conf"
)

// reSpaces matches more than one consecutive spaces.
// This is used to split keys from values
// from the text config.
var reSpaces = regexp.MustCompile(`\s+`)

// Parse a classic fluent-bit config.
func Parse(text string) (fluentbitconf.Config, error) {
	var out fluentbitconf.Config

	c, err := parse(text)
	if err != nil {
		return out, err
	}

	addByName := func(props Properties, dest *[]fluentbitconf.ByName) {
		if nameVal, ok := props.Get("Name"); ok {
			if name, ok := nameVal.(string); ok {
				config := fluentbitconf.Properties{}
				for _, p := range props {
					// Case when the property could be repeated. Example:
					// 	[FILTER]
					// 		Name record_modifier
					// 		Match *
					// 		Record hostname ${HOSTNAME}
					// 		Record product Awesome_Tool
					if v, ok := config.Get(p.Name); ok {
						if s, ok := v.([]any); ok {
							s = append(s, p.Value)
							config.Set(p.Name, s)
						} else {
							config.Set(p.Name, []any{v, p.Value})
						}
					} else {
						config.Set(p.Name, p.Value)
					}
				}
				*dest = append(*dest, fluentbitconf.ByName{
					name: config,
				})
			}
		}
	}

	for _, entry := range c.Entries {
		switch entry.Kind {
		case EntryKindCommand:
			if strings.EqualFold(entry.AsCommand.Name, "SET") {
				if out.Env == nil {
					out.Env = fluentbitconf.Properties{}
				}
				out.Env.Set(splitCommand(entry.AsCommand.Instruction))
			}
		case EntryKindSection:
			switch {
			case strings.EqualFold(entry.AsSection.Name, "SERVICE"):
				if out.Service == nil {
					out.Service = fluentbitconf.Properties{}
				}
				for _, p := range entry.AsSection.Properties {
					out.Service.Set(p.Name, p.Value)
				}
			case strings.EqualFold(entry.AsSection.Name, "CUSTOM"):
				addByName(entry.AsSection.Properties, &out.Customs)
			case strings.EqualFold(entry.AsSection.Name, "INPUT"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Inputs)
			case strings.EqualFold(entry.AsSection.Name, "FILTER"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Filters)
			case strings.EqualFold(entry.AsSection.Name, "OUTPUT"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Outputs)
			}
		}
	}

	return out, nil
}

func parse(text string) (Config, error) {
	var config Config

	lines := strings.Split(text, "\n")

	var currentSection *Section

	flushCurrentSection := func() {
		if currentSection == nil {
			return
		}

		sec := *currentSection
		config.Entries = append(config.Entries, Entry{
			Kind:      EntryKindSection,
			AsSection: &sec,
		})

		currentSection = nil
	}

	for i, line := range lines {
		lineNumber := uint(i + 1)
		if !utf8.ValidString(line) {
			return config, NewError("invalid utf8 string", lineNumber)
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

			if err := config.addCommand(line); err != nil {
				return config, NewError(err.Error(), lineNumber)
			}

			continue
		}

		// beginning of section
		if strings.HasPrefix(line, "[") {
			flushCurrentSection()

			if !strings.HasSuffix(line, "]") {
				return config, NewError("expected section to end with \"]\"", lineNumber)
			}

			sec, err := newSection(line)
			if err != nil {
				return config, NewError(err.Error(), lineNumber)
			}

			currentSection = &sec

			continue
		}

		if currentSection == nil {
			return config, NewError(fmt.Sprintf("unexpected entry %q; expected beginning of section", line), lineNumber)
		}

		// property within section
		if err := addSectionProperty(currentSection, line); err != nil {
			return config, NewError(err.Error(), lineNumber)
		}
	}

	flushCurrentSection()

	return config, nil
}

func newSection(line string) (Section, error) {
	sectionName := strings.Trim(line, "[]")
	sectionName = strings.TrimSpace(sectionName)

	if sectionName == "" {
		return Section{}, errors.New("expected section name to not be empty")
	}

	return Section{
		Name: sectionName,
	}, nil
}

func addSectionProperty(sec *Section, line string) error {
	name, value, err := splitKeyVal(line)
	if err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	if sec.Properties == nil {
		sec.Properties = Properties{}
	}
	sec.Properties.Add(name, anyFromString(value))

	return nil
}

func (c *Config) addCommand(line string) error {
	line = strings.TrimPrefix(line, "@")
	name, instruct, err := splitKeyVal(line)
	if err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	c.Entries = append(c.Entries, Entry{
		Kind: EntryKindCommand,
		AsCommand: &Command{
			Name:        name,
			Instruction: instruct,
		},
	})

	return nil
}

func splitKeyVal(s string) (string, string, error) {
	parts := reSpaces.Split(s, 2)
	if len(parts) != 2 {
		return "", "", errors.New("expected at least two strings separated by a space")
	}
	return parts[0], parts[1], nil
}

func splitCommand(text string) (string, string) {
	name, val, _ := strings.Cut(text, "=")
	name = strings.TrimSpace(name)
	val = strings.TrimSpace(val)
	return name, val
}
