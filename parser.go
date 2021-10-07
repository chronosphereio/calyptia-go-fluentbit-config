package fluentbit_config

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

var (
	DefaultLexerRules = []lexer.Rule{
		{"DateTime", `\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?(-\d\d:\d\d)?`, nil},
		{"Date", `\d\d\d\d-\d\d-\d\d`, nil},
		{"Time", `\d\d:\d\d:\d\d(\.\d+)?`, nil},
		{"Ident", `[a-zA-Z_\.\*\-0-9\/]+`, nil},
		{"String", `[a-zA-Z0-9_\.\/\*\-]+`, nil},
		{"Number", `[-+]?[.0-9]+\b`, nil},
		{`Float`, `(\d*)(\.)*(\d+)+`, nil},
		{"Punct", `\[|]|[-!()+/*=,]`, nil},
		{"comment", `#[^\n]+`, nil},
		{"whitespace", `\s+`, nil},
		{"EOL", "[\n]+", nil},
	}
)

type ConfigGrammar struct {
	Pos     lexer.Position
	Entries []*Entry `@@*`
}

type Entry struct {
	Field   *Field   `  @@`
	Section *Section `| @@`
}

type Section struct {
	Name   string   `"[" @(Ident ( "." Ident )*) "]"`
	Fields []*Field `@@*`
}

type Field struct {
	Key   string `@Ident`
	Value *Value `@@`
}

type Value struct {
	String   *string  ` @Ident`
	DateTime *string  `| @DateTime`
	Date     *string  `| @Date`
	Time     *string  `| @Time`
	Bool     *bool    `| (@"true" | "false")`
	Number   *float64 `| @Number`
	Float    *float64 `| @Float`
	List     []*Value `| "[" ( @@ ( "," @@ )* )? "]"`
}

type Config struct {
	Inputs  map[string][]Field
	Filters map[string][]Field
	Customs map[string][]Field
	Outputs map[string][]Field
}

func addFields(e *Entry, m *map[string][]Field) {
	if *m == nil {
		*m = map[string][]Field{}
	}
	q := *m
	var name string
	for _, field := range e.Section.Fields {
		if strings.ToLower(field.Key) == "name" {
			name = *field.Value.String
		}
	}

	for _, field := range e.Section.Fields {
		q[name] = append(q[name], *field)
	}
}

func (c *Config) loadSectionsFromGrammar(grammar *ConfigGrammar) error {
	for _, entry := range grammar.Entries {
		switch entry.Section.Name {
		case "INPUT":
			{
				addFields(entry, &c.Inputs)
			}
		case "FILTER":
			{
				addFields(entry, &c.Filters)
			}
		case "OUTPUT":
			{
				addFields(entry, &c.Outputs)
			}
		case "CUSTOM":
			{
				addFields(entry, &c.Customs)
			}
		}
	}
	return nil
}

func NewFromBytes(data []byte) (*Config, error) {
	var grammar = &ConfigGrammar{
		Entries: []*Entry{},
	}
	var config Config

	statefulDefinition, err := lexer.NewSimple(DefaultLexerRules)
	if err != nil {
		return nil, err
	}

	parser := participle.MustBuild(
		grammar,
		participle.Lexer(
			lexer.Must(statefulDefinition, err),
		),
	)

	err = parser.ParseBytes("", data, grammar)
	if err != nil {
		return nil, err
	}

	if len(grammar.Entries) == 0 {
		return nil, fmt.Errorf("no configuration entries found in provided grammar")
	}

	err = config.loadSectionsFromGrammar(grammar)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
