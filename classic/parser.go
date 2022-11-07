package classic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/calyptia/go-fluentbit-config/property"
)

// reSpaces matches more than one consecutive spaces.
// This is used to split keys from values
// from the text config.
var reSpaces = regexp.MustCompile(`\s+`)

func (c *Classic) UnmarshalText(text []byte) error {
	lines := strings.Split(string(text), "\n")

	var currentSection *Section

	flushCurrentSection := func() {
		if currentSection == nil {
			return
		}

		sec := *currentSection
		c.Entries = append(c.Entries, Entry{
			Kind:      EntryKindSection,
			AsSection: &sec,
		})

		currentSection = nil
	}

	for i, line := range lines {
		lineNumber := uint(i + 1)
		if !utf8.ValidString(line) {
			return NewError("invalid utf8 string", lineNumber)
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

			cmd, err := newCommand(line)
			if err != nil {
				return NewError(err.Error(), lineNumber)
			}

			c.Entries = append(c.Entries, Entry{
				Kind:      EntryKindCommand,
				AsCommand: cmd,
			})

			continue
		}

		// beginning of section
		if strings.HasPrefix(line, "[") {
			flushCurrentSection()

			if !strings.HasSuffix(line, "]") {
				return NewError("expected section to end with \"]\"", lineNumber)
			}

			var err error
			currentSection, err = newSection(line)
			if err != nil {
				return NewError(err.Error(), lineNumber)
			}

			continue
		}

		if currentSection == nil {
			return NewError(fmt.Sprintf("unexpected entry %q", line), lineNumber)
		}

		// property within section
		if err := addSectionProperty(currentSection, line); err != nil {
			return NewError(err.Error(), lineNumber)
		}
	}

	flushCurrentSection()

	return nil
}

func (c Classic) MarshalText() ([]byte, error) {
	var sb strings.Builder

	for _, entry := range c.Entries {
		switch entry.Kind {
		case EntryKindCommand:
			_, err := fmt.Fprintf(&sb, "@%s %s\n", entry.AsCommand.Name, entry.AsCommand.Instruction)
			if err != nil {
				return nil, err
			}
		case EntryKindSection:
			_, err := fmt.Fprintf(&sb, "[%s]\n", entry.AsSection.Name)
			if err != nil {
				return nil, err
			}
			tw := tabwriter.NewWriter(&sb, 0, 4, 1, ' ', 0)
			for _, p := range entry.AsSection.Properties {
				if s, ok := p.Value.([]any); ok {
					for _, v := range s {
						_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, stringFromAny(v))
						if err != nil {
							return nil, err
						}
					}
				} else {
					_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, stringFromAny(p.Value))
					if err != nil {
						return nil, err
					}
				}
			}
			err = tw.Flush()
			if err != nil {
				return nil, err
			}
		}
	}

	return []byte(sb.String()), nil
}

func newSection(line string) (*Section, error) {
	sectionName := strings.Trim(line, "[]")
	sectionName = strings.TrimSpace(sectionName)

	if sectionName == "" {
		return nil, errors.New("expected section name to not be empty")
	}

	return &Section{
		Name: sectionName,
	}, nil
}

func addSectionProperty(sec *Section, line string) error {
	name, value, err := splitKeyVal(line)
	if err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	if sec.Properties == nil {
		sec.Properties = property.Properties{}
	}
	sec.Properties.Add(name, anyFromString(value))

	return nil
}

func newCommand(line string) (*Command, error) {
	line = strings.TrimPrefix(line, "@")
	name, instruct, err := splitKeyVal(line)
	if err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	return &Command{
		Name:        name,
		Instruction: instruct,
	}, nil
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
