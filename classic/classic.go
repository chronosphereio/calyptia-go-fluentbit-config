package classic

import "strings"

// Config grammar of the classic fluent-bit config files.
// It resembles to INI config format files.
type Config struct {
	Entries []Entry
}

// Entry within a config. It can be either a command, or a section.
// Comments are ignored.
type Entry struct {
	Kind      EntryKind
	AsCommand *Command
	AsSection *Section
}

// EntryKind enum can be either command or section.
type EntryKind string

const (
	EntryKindCommand EntryKind = "command"
	EntryKindSection EntryKind = "section"
)

// Command within a config.
// Known commands are SET and INCLUDE.
type Command struct {
	Name        string
	Instruction string
}

// Section Section within a config.
type Section struct {
	Name       string
	Properties Properties
}

type Properties []Property

func (pp Properties) Has(name string) bool {
	if pp == nil {
		return false
	}

	for _, p := range pp {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}

	return false
}

func (pp Properties) Get(name string) (any, bool) {
	if pp == nil {
		return nil, false
	}

	for _, p := range pp {
		if strings.EqualFold(p.Name, name) {
			return p.Value, true
		}
	}

	return nil, false
}

func (pp *Properties) Set(name string, value any) {
	if pp == nil {
		return
	}

	for i, p := range *pp {
		if strings.EqualFold(p.Name, name) {
			(*pp)[i].Value = value
			return
		}
	}

	pp.Add(name, value)
}

func (pp *Properties) Add(name string, value any) {
	if pp == nil {
		return
	}

	*pp = append(*pp, Property{Name: name, Value: value})
}

// Property nested within a section inside a config.
type Property struct {
	Name string
	// Value can be a bool, float64, string or a slice of those.
	Value any
}
