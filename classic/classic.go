package classic

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

// Section section within a config.
type Section struct {
	Name       string
	Properties []Property
}

// Property nested within a section inside a config.
type Property struct {
	Name  string
	Value Value
}
