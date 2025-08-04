package fluentbitconfig

type SectionKind string

func (k SectionKind) String() string { return string(k) }

const (
	SectionKindService   SectionKind = "service"
	SectionKindCustom    SectionKind = "custom"
	SectionKindInput     SectionKind = "input"
	SectionKindParser    SectionKind = "parser"
	SectionKindFilter    SectionKind = "filter"
	SectionKindOutput    SectionKind = "output"
	SectionKindProcessor SectionKind = "processor"
)
