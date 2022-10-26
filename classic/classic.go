package classic

import "github.com/calyptia/go-fluentbit-conf/property"

// Classic grammar of the legacy fluent-bit config files.
// It resembles to INI config format files.
type Classic struct {
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
// These begin with an `@` symbol.
// Example:
//
//	@SET foo=bar
//	@INCLUDE /some/path/file.ext
type Command struct {
	Name        string
	Instruction string
}

// Section within a config.
// These start with a section name between square brackets
// followed by a list of nested properties.
// Each property is separated by a space.
// Example:
//
//	[INPUT]
//		Name dummy
//	[FILTER]
//		Name  record_modifier
//		Match *
//	[OUTPUT]
//		Name  stdout
//		Match *
type Section struct {
	Name       string
	Properties property.Properties
}
