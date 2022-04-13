package fluentbit_config

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	DefaultLexerRules = []lexer.Rule{
		{"DateTime", `\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?(-\d\d:\d\d)?`, nil},
		{"Date", `\d\d\d\d-\d\d-\d\d`, nil},
		{"Time", `\d\d:\d\d:\d\d(\.\d+)?`, nil},
		{"TimeFormat", `^((%[dbYHMSz])(\:|\/|\ |))+$`, nil},
		{"Ident", `[a-zA-Z_\.\*\-0-9\/\{\}]+`, nil},
		{"String", `[a-zA-Z0-9_\.\/\*\-]+`, nil},
		{"Number", `[-+]?[.0-9]+\b`, nil},
		{`Float`, `(\d*)(\.)*(\d+)+`, nil},
		{"Punct", `\[|]|[-!()+/*=,]`, nil},
		{"Regex", `\^[^\n]+\$`, nil},
		{"Topic", `\$[a-zA-Z0-9_\.\/\*\-]+`, nil},
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
	String     *string  ` @Ident`
	DateTime   *string  `| @DateTime`
	Date       *string  `| @Date`
	Time       *string  `| @Time`
	TimeFormat *string  `| @TimeFormat`
	Topic      *string  `| @Topic`
	Regex      *string  `| @Regex`
	Bool       *bool    `| (@"true" | "false")`
	Number     *float64 `| @Number`
	Float      *float64 `| @Float`
	List       []*Value `| "[" ( @@ ( "," @@ )* )? "]"`
}

type Config struct {
	Inputs      map[string][]Field
	Filters     map[string][]Field
	Customs     map[string][]Field
	Outputs     map[string][]Field
	Parsers     map[string][]Field
	InputIndex  int
	FilterIndex int
	OutputIndex int
	CustomIndex int
	ParserIndex int
}

func addFields(e *Entry, index int, m *map[string][]Field) {
	if *m == nil {
		*m = map[string][]Field{}
	}
	q := *m
	var name string
	for _, field := range e.Section.Fields {
		if strings.ToLower(field.Key) == "name" {
			name = fmt.Sprintf("%s.%d", *field.Value.String, index)
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
				addFields(entry, c.InputIndex, &c.Inputs)
				c.InputIndex++
			}
		case "FILTER":
			{
				addFields(entry, c.FilterIndex, &c.Filters)
				c.FilterIndex++
			}
		case "OUTPUT":
			{
				addFields(entry, c.OutputIndex, &c.Outputs)
				c.OutputIndex++
			}
		case "CUSTOM":
			{
				addFields(entry, c.CustomIndex, &c.Customs)
				c.CustomIndex++
			}
		case "PARSER":
			{
				addFields(entry, c.ParserIndex, &c.Parsers)
				c.ParserIndex++
			}
		}
	}
	return nil
}

func NewFromBytes(data []byte, skipProperties ...string) (*Config, error) {
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
