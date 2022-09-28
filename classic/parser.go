package classic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// reSpaces matches more than one consecutive spaces.
// This is used to split keys from values
// from the text config.
var reSpaces = regexp.MustCompile(`\s+`)

// Parse a classic fluent-bit config.
func Parse(text string) (Config, error) {
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

	sec.Properties = append(sec.Properties, Property{
		Name:  name,
		Value: valueFromString(value),
	})
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
	switch len(parts) {
	case 0, 1:
		return "", "", errors.New("expected at least two strings separated by a space")
	}
	return parts[0], parts[1], nil
}
