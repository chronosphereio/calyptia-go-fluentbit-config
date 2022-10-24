package fluentbit_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gopkg.in/yaml.v2"
)

var (
	DefaultLexerRules = []lexer.SimpleRule{
		{"Include", `@[Ii][Nn][Cc][Ll][Uu][Dd][Ee]`},
		{"Set", `@[Ss][Ee][Tt]`},
		{"TemplateVariable", `\{\{[^\}]+\}\}`},
		{"JsonObject", `\{[^\}]+\}`},
		{"DateTime", `\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d(\.\d+)?`},
		{"Date", `\d\d\d\d-\d\d-\d\d`},
		{"Time", `\d\d:\d\d:\d\d(\.\d+)?`},
		{"TimeFormat", `((%([aAbBhcCdeDHIjmMnprRStTUwWxXyYLz]|E[cCxXyY]|O[deHImMSUwWy]))(\:|\/| |\-|T|Z|\.|))+`},
		{"Address", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`},
		{"Unit", `[-+]?(\d*\.)?\d+([kKmMgG])`},
		{"Float", `[-+]?\d*\.\d+`},
		{"UUID", `^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`},
		{"Number", `[-+]?\d+`},
		{"Ident", `[a-zA-Z_\.\*\-0-9]+`},
		{"List", `([a-zA-Z_\.\*\-0-9\/]+ )+[a-zA-Z_\.\*\-0-9\/]+`},
		{"String", `[a-zA-Z_\,\.\:\*\-0-9\/\$\%\\\/]+`},
		{"Punct", `\[|]|[-!()+/*=,]`},
		{"Regex", `\^[^\n]+\$`},
		{"Topic", `\$[a-zA-Z0-9_\.\/\*\-]+`},
		{"comment", `#[^\n]+`},
		{"whitespace", `[ \t]+`},
		{"EOL", "[\r\n]*"},
	}
)

type ConfigFormat string

const (
	ConfigFormatINI  ConfigFormat = "ini"
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatYAML ConfigFormat = "yaml"
)

type ConfigGrammar struct {
	Pos     lexer.Position
	Start   string   `@EOL*`
	Entries []*Entry `@@*`
	End     string   `@EOL*`
}

type Entry struct {
	Include *Include `  @@`
	Set     *Set     `| @@`
	Section *Section `| @@`
}

type Set struct {
	Declaration string ` @Set`
	Lval        string ` (@Ident | @String) "="`
	Rval        string ` @Ident | @String`
}

type Include struct {
	Declaration string ` @Include`
	Include     string ` @Ident | @String`
}

type Section struct {
	Name    string   `"[" @(Ident ( "." Ident )*) "]"`
	Padding string   ` @EOL*`
	Fields  []*Field `@@*`
}

type Field struct {
	Key    string  `@Ident`
	Values Values  ` @@*`
	End    *string ` @EOL*`
}

type Value struct {
	JsonObject       *string  `@JsonObject`
	TemplateVariable *string  `| @TemplateVariable`
	Uuid             *string  `| @UUID`
	Address          *string  `| @Address`
	DateTime         *string  `| @DateTime`
	Date             *string  `| @Date`
	Time             *string  `| @Time`
	TimeFormat       *string  `| @TimeFormat`
	Unit             *string  `| @Unit`
	Number           *int64   `| @Number`
	Float            *float64 `| @Float`
	Topic            *string  `| @Topic`
	Regex            *string  `| @Regex`
	String           *string  `| @Ident | @String`
	Bool             *bool    `| (@"true" | "false")`
	List             []string `| @List`
}

type Values []Value

func (vs Values) ToString() string {
	vals := make([]string, 0)

	for _, val := range vs {
		vals = append(vals, val.ToString())
	}

	return strings.Join(vals, " ")
}

func (fields Fields) hasField(name string) bool {
	for _, field := range fields {
		if strings.EqualFold(name, field.Key) {
			return true
		}
	}
	return false
}

func (a *Value) Equals(b *Value) bool {
	if a.JsonObject != nil {
		if b.JsonObject == nil {
			return false
		}
		return *a.JsonObject == *b.JsonObject
	} else if a.TemplateVariable != nil {
		if b.TemplateVariable == nil {
			return false
		}
		return *a.TemplateVariable == *b.TemplateVariable
	} else if a.String != nil {
		if b.String == nil {
			return false
		}
		return *a.String == *b.String
	} else if a.DateTime != nil {
		if b.DateTime == nil {
			return false
		}
		return *a.DateTime == *b.DateTime
	} else if a.Date != nil {
		if b.Date == nil {
			return false
		}
		return *a.Date == *b.Date
	} else if a.Time != nil {
		if b.Time == nil {
			return false
		}
		return *a.Time == *b.Time
	} else if a.TimeFormat != nil {
		if b.TimeFormat == nil {
			return false
		}
		return *a.TimeFormat == *b.TimeFormat
	} else if a.Topic != nil {
		if b.Topic == nil {
			return false
		}
		return *a.Topic == *b.Topic
	} else if a.Regex != nil {
		if b.Regex == nil {
			return false
		}
		return *a.Regex == *b.Regex
	} else if a.Bool != nil {
		if b.Bool == nil {
			return false
		}
		return *a.Bool == *b.Bool
	} else if a.Number != nil {
		if b.Number == nil {
			return false
		}
		return *a.Number == *b.Number
	} else if a.Float != nil {
		if b.Float == nil {
			return false
		}
		return *a.Float == *b.Float
	} else if a.Address != nil {
		if b.Address == nil {
			return false
		}
		return *a.Address == *b.Address
	}
	return false
}

func (a *Value) Value() interface{} {
	switch {
	case a.Address != nil:
		return *a.Address
	case a.TemplateVariable != nil:
		return *a.TemplateVariable
	case a.JsonObject != nil:
		return *a.JsonObject
	case a.String != nil:
		return *a.String
	case a.DateTime != nil:
		return *a.DateTime
	case a.Date != nil:
		return *a.Date
	case a.Time != nil:
		return *a.Time
	case a.TimeFormat != nil:
		return *a.TimeFormat
	case a.Topic != nil:
		return *a.Topic
	case a.Regex != nil:
		return *a.Regex
	case a.Bool != nil:
		return *a.Bool
	case a.Number != nil:
		return *a.Number
	case a.Float != nil:
		return *a.Float
	}
	return nil
}

func (a *Value) ToString() string {
	val := a.Value()
	switch t := val.(type) {
	case string:
		return t
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		return fmt.Sprintf("%0.6f", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	}
	return ""
}

type ConfigSectionType int

const (
	NoneSection ConfigSectionType = iota
	ServiceSection
	InputSection
	FilterSection
	OutputSection
	CustomSection
	ParserSection
	IncludeSection
	SetSection
)

var ConfigSectionTypes []string = []string{
	"NONE",
	"SERVICE",
	"INPUT",
	"FILTER",
	"OUTPUT",
	"CUSTOM",
	"PARSER",
	"INCLUDE",
	"SET",
}

type ConfigSection struct {
	Type   ConfigSectionType
	ID     string
	Fields []Field
}

type Config struct {
	PluginIndex map[string]int
	Sections    []ConfigSection
}

func ptr[T any](v T) *T {
	return &v
}

func (c *Config) addSection(sectype ConfigSectionType, e *Entry) {
	var plugin string
	var ok bool
	section := ConfigSection{Type: sectype}

	for _, field := range e.Section.Fields {
		if strings.ToLower(field.Key) == "name" {
			plugin = field.Values.ToString()
		}
		section.Fields = append(section.Fields, *field)
	}

	index, ok := c.PluginIndex[plugin]
	if !ok {
		index = 0
	}
	section.ID = fmt.Sprintf("%s.%d", plugin, index)
	c.PluginIndex[plugin] = index + 1

	c.Sections = append(c.Sections, section)
}

func (c *Config) addInclude(include string) {
	c.Sections = append(c.Sections, ConfigSection{
		Type: IncludeSection,
		Fields: []Field{{
			Key: "@include",
			Values: []Value{{
				String: ptr(include),
			}},
		}},
	})
}

func (c *Config) addSet(lval, rval string) {
	c.Sections = append(c.Sections, ConfigSection{
		Type: SetSection,
		Fields: []Field{{
			Key: lval,
			Values: []Value{{
				String: ptr(rval),
			}},
		}},
	})
}

func (c *Config) loadSectionsFromGrammar(grammar *ConfigGrammar) error {
	for _, entry := range grammar.Entries {
		if entry.Section != nil {
			switch entry.Section.Name {
			case "SERVICE":
				{
					c.addSection(ServiceSection, entry)
				}
			case "INPUT":
				{
					c.addSection(InputSection, entry)
				}
			case "FILTER":
				{
					c.addSection(FilterSection, entry)
				}
			case "OUTPUT":
				{
					c.addSection(OutputSection, entry)
				}
			case "CUSTOM":
				{
					c.addSection(CustomSection, entry)
				}
			case "PARSER":
				{
					c.addSection(ParserSection, entry)
				}
			}
		} else if entry.Include != nil {
			c.addInclude(entry.Include.Include)
		} else if entry.Set != nil {
			c.addSet(entry.Set.Lval, entry.Set.Rval)
		}
	}
	return nil
}

func NewFromBytes(data []byte, skipProperties ...string) (*Config, error) {
	if len(skipProperties) > 0 {
		fmt.Println("deprecated: we no longer support skipping properties")
	}
	return ParseINI(data)
}

func (a *Value) dumpValueINI(ini *bytes.Buffer) error {
	// ... this might be better served with an interface
	// and type switching
	// type Value struct {
	// 	String     *string  ` @Ident`
	// 	DateTime   *string  `| @DateTime`
	// 	Date       *string  `| @Date`
	// 	Time       *string  `| @Time`
	// 	TimeFormat *string  `| @TimeFormat`
	// 	Topic      *string  `| @Topic`
	// 	Regex      *string  `| @Regex`
	// 	Bool       *bool    `| (@"true" | "false")`
	// 	Number     *float64 `| @Number`
	// 	Float      *float64 `| @Float`
	// 	List       []*Value `| "[" ( @@ ( "," @@ )* )? "]"`
	// }
	switch true {
	case a.String != nil:
		ini.Write([]byte(*a.String))
	case a.DateTime != nil:
		ini.Write([]byte(*a.DateTime))
	case a.Date != nil:
		ini.Write([]byte(*a.Date))
	case a.Time != nil:
		ini.Write([]byte(*a.Time))
	case a.TimeFormat != nil:
		ini.Write([]byte(*a.TimeFormat))
	case a.Topic != nil:
		ini.Write([]byte(fmt.Sprintf("$%s", *a.Topic)))
	case a.Regex != nil:
		ini.Write([]byte(*a.Regex))
	case a.Bool != nil:
		if *a.Bool {
			ini.Write([]byte("on"))
		} else {
			ini.Write([]byte("off"))
		}
	case a.Number != nil:
		ini.Write([]byte(fmt.Sprintf("%d", int(*a.Number))))
	case a.Float != nil:
		ini.Write([]byte(fmt.Sprintf("%0.6f", *a.Float)))
	case a.JsonObject != nil:
		ini.Write([]byte(*a.JsonObject))
	case a.Address != nil:
		ini.Write([]byte(*a.Address))
	case a.List != nil:
		//for _, v := range value.List {
		//if err := v.dumpValueINI(ini); err != nil {
		//	return err
		//}
		//}
	}
	return nil
}

func (vs Values) dumpValueINI(ini *bytes.Buffer) error {
	for _, value := range vs {
		if err := value.dumpValueINI(ini); err != nil {
			return err
		}
	}
	return nil
}

func (f *Field) dumpFieldINI(ini *bytes.Buffer) error {
	ini.Write([]byte(fmt.Sprintf("    %s ", f.Key)))
	if err := f.Values.dumpValueINI(ini); err != nil {
		return err
	}
	if f.End != nil {
		ini.Write([]byte(*f.End))
	} else {
		ini.Write([]byte("\n"))
	}
	return nil
}

func (c *Config) DumpAs(configFormat ConfigFormat) ([]byte, error) {
	switch configFormat {
	case ConfigFormatYAML:
		return DumpYAML(c)
	case ConfigFormatJSON:
		return DumpJSON(c)
	case ConfigFormatINI:
		return DumpINI(c)
	}
	return nil, fmt.Errorf("cannot dump config to unknown format: %s", configFormat)
}

func (c *Config) DumpINI() ([]byte, error) {
	return DumpINI(c)
}

func DumpINI(c *Config) ([]byte, error) {
	ini := bytes.NewBuffer([]byte(""))

	for _, section := range c.Sections {
		switch section.Type {
		case ServiceSection:
			fallthrough
		case InputSection:
			fallthrough
		case FilterSection:
			fallthrough
		case OutputSection:
			fallthrough
		case ParserSection:
			fallthrough
		case CustomSection:
			ini.Write([]byte(fmt.Sprintf("[%s]\n",
				ConfigSectionTypes[section.Type])))

			for _, field := range section.Fields {
				if err := field.dumpFieldINI(ini); err != nil {
					return []byte(""), err
				}
			}
		case IncludeSection:
			if len(section.Fields) != 1 {
				return []byte(""), fmt.Errorf("invalid input section")
			}
			if len(section.Fields[0].Values) != 1 {
				return []byte(""), fmt.Errorf("invalid input section")
			}
			if section.Fields[0].Values[0].String == nil {
				return []byte(""), fmt.Errorf("invalid input section")
			}
			ini.Write([]byte(fmt.Sprintf("@include %s\n",
				*section.Fields[0].Values[0].String)))
		case SetSection:
			for _, field := range section.Fields {
				for _, value := range field.Values {
					ini.Write([]byte(
						fmt.Sprintf("@set %s = %s\n", field.Key, value.ToString())))
				}
			}
		}
	}

	return ini.Bytes(), nil
}

func ParseINI(data []byte) (*Config, error) {
	var grammar = &ConfigGrammar{
		Entries: []*Entry{},
	}
	var config = Config{
		PluginIndex: make(map[string]int),
	}

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

type Fields []Field

func (fields *Fields) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	kv := make(map[string]interface{})
	if err := unmarshal(&kv); err != nil {
		return err
	}
	newFields := make([]Field, 0)
	for k, v := range kv {
		switch t := v.(type) {
		case string:
			newFields = append(newFields, Field{
				Key: k,
				Values: []Value{{
					String: &t,
				}},
			})
		case int:
			i := int64(v.(int))
			newFields = append(newFields, Field{
				Key: k,
				Values: []Value{{
					Number: &i,
				}},
			})
		case int64:
			newFields = append(newFields, Field{
				Key: k,
				Values: []Value{{
					Number: &t,
				}},
			})
		case float64:
			newFields = append(newFields, Field{
				Key: k,
				Values: []Value{{
					Float: &t,
				}},
			})
		case bool:
			newFields = append(newFields, Field{
				Key: k,
				Values: []Value{{
					Bool: &t,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v: %+v", t, v)
		}
	}
	*fields = newFields

	return nil
}

func (fields Fields) MarshalYAML() (interface{}, error) {
	fmap := make(map[string]interface{})
	for _, field := range fields {
		if len(field.Values) == 1 {
			switch true {
			case field.Values[0].String != nil:
				fmap[field.Key] = *field.Values[0].String
			case field.Values[0].DateTime != nil:
				fmap[field.Key] = *field.Values[0].DateTime
			case field.Values[0].Date != nil:
				fmap[field.Key] = *field.Values[0].Date
			case field.Values[0].Time != nil:
				fmap[field.Key] = *field.Values[0].Time
			case field.Values[0].Bool != nil:
				fmap[field.Key] = *field.Values[0].Bool
			case field.Values[0].Number != nil:
				fmap[field.Key] = *field.Values[0].Number
			case field.Values[0].Float != nil:
				fmap[field.Key] = *field.Values[0].Float
			default:
				return nil, fmt.Errorf("unknown type: %+v", field)
			}
		} else {
			fmap[field.Key] = field.Values.ToString()
		}
	}
	return fmap, nil
}

func (fields Fields) MarshalJSON() ([]byte, error) {
	bfields := bytes.NewBuffer([]byte(""))
	bfields.Write([]byte("{"))

	for fidx, field := range fields {
		bfields.Write([]byte(fmt.Sprintf("\"%s\": ", field.Key)))
		val := bytes.NewBuffer([]byte(""))
		isstr := false

		if len(field.Values) == 1 {
			switch true {
			case field.Values[0].String != nil:
				if isstr {
					val.Write([]byte(" "))
				} else {
					isstr = true
				}
				val.Write([]byte(*field.Values[0].String))
			case field.Values[0].DateTime != nil:
				if isstr {
					val.Write([]byte(" "))
				} else {
					isstr = true
				}
				val.Write([]byte(*field.Values[0].DateTime))
			case field.Values[0].Date != nil:
				if isstr {
					val.Write([]byte(" "))
				} else {
					isstr = true
				}
				val.Write([]byte(*field.Values[0].Date))
			case field.Values[0].Time != nil:
				if isstr {
					val.Write([]byte(" "))
				} else {
					isstr = true
				}
				val.Write([]byte(*field.Values[0].Time))
			case field.Values[0].Bool != nil:
				if *field.Values[0].Bool {
					val.Reset()
					val.Write([]byte("true"))
				} else {
					val.Reset()
					val.Write([]byte("false"))
				}
			case field.Values[0].Number != nil:
				val.Write([]byte(fmt.Sprintf("%d", int(*field.Values[0].Number))))
			case field.Values[0].Float != nil:
				val.Write([]byte(fmt.Sprintf("%0.6f", *field.Values[0].Float)))
			case field.Values[0].JsonObject != nil:
				val.Write([]byte(strconv.Quote(*field.Values[0].JsonObject)))
			case field.Values[0].TemplateVariable != nil:
				val.Write([]byte(strconv.Quote(*field.Values[0].TemplateVariable)))
			case field.Values[0].Address != nil:
				val.Write([]byte(strconv.Quote(*field.Values[0].Address)))
			case field.Values[0].Unit != nil:
				val.Write([]byte(strconv.Quote(*field.Values[0].Unit)))
			default:
				return nil, fmt.Errorf("unknown typed: %+v", field)
			}
		} else {
			isstr = true
			val.Write([]byte(field.Values.ToString()))
		}

		if isstr {
			bfields.Write([]byte(strconv.Quote(val.String())))
		} else {
			bfields.Write(val.Bytes())
		}

		if fidx < len(fields)-1 {
			bfields.Write([]byte(", "))
		}
	}

	bfields.Write([]byte("}"))
	return bfields.Bytes(), nil
}

func (fields *Fields) UnmarshalJSON(b []byte) error {
	newFields := make(Fields, 0)

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		return err
	}

	if t != json.Delim('{') {
		return fmt.Errorf(
			"unable to find starting field delimiters, found instead: %+v (%d)",
			t, dec.InputOffset())
	}
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			return err
		}
		var v interface{}
		if err := dec.Decode(&v); err != nil {
			return err
		}

		switch t := v.(type) {
		case string:
			newFields = append(newFields, Field{
				Key: k.(string),
				Values: []Value{{
					String: &t,
				}},
			})
		case json.Number:
			n := v.(json.Number)
			if i, err := n.Int64(); err != nil {
				f, err := n.Float64()
				if err != nil {
					return err
				}
				newFields = append(newFields, Field{
					Key: k.(string),
					Values: []Value{{
						Float: &f,
					}},
				})
			} else {
				newFields = append(newFields, Field{
					Key: k.(string),
					Values: []Value{{
						Number: &i,
					}},
				})
			}
		case bool:
			newFields = append(newFields, Field{
				Key: k.(string),
				Values: []Value{{
					Bool: &t,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v", v)
		}
	}

	*fields = newFields
	return nil
}

type yamlGrammarPipeline struct {
	Inputs  []map[string]Fields `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Filters []map[string]Fields `yaml:"filters,omitempty" json:"filters,omitempty"`
	Parsers []map[string]Fields `yaml:"parsers,omitempty" json:"parsers,omitempty"`
	Outputs []map[string]Fields `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

type yamlGrammar struct {
	Service  Fields              `yaml:"service,omitempty" json:"service,omitempty"`
	Customs  []map[string]Fields `yaml:"customs,omitempty" json:"customs,omitempty"`
	Pipeline yamlGrammarPipeline `yaml:"pipeline" json:"pipeline"`
	Includes []string            `yaml:"includes,omitempty" json:"includes,omitempty"`
}

func getPluginName(fields map[string]Fields) string {
	rkey := ""
	for key := range fields {
		rkey = key
	}
	return rkey
}

func getPluginNameParameter(cfg ConfigSection) string {
	for _, field := range cfg.Fields {
		if strings.ToLower(field.Key) == "name" {
			return field.Values.ToString()
		}
	}
	return ""
}

func (c *Config) dumpYamlGrammar() (*yamlGrammar, error) {
	yg := yamlGrammar{
		Service: make(Fields, 0),
		Customs: make([]map[string]Fields, 0),
		Pipeline: yamlGrammarPipeline{
			Inputs:  make([]map[string]Fields, 0),
			Filters: make([]map[string]Fields, 0),
			Outputs: make([]map[string]Fields, 0),
			Parsers: make([]map[string]Fields, 0),
		},
		Includes: make([]string, 0),
	}

	for _, section := range c.Sections {
		if section.Type == ServiceSection {
			for _, field := range section.Fields {
				yg.Service = append(yg.Service, field)
			}
		} else if section.Type == IncludeSection {
			yg.Includes = append(yg.Includes,
				*section.Fields[0].Values[0].String)

		} else if section.Type == SetSection {
			return nil, fmt.Errorf("@set is unsupported")
		} else {
			sectionName := getPluginNameParameter(section)
			sectionmap := make(map[string]Fields)
			sectionmap[sectionName] = make(Fields, 0)

			for _, field := range section.Fields {
				sectionmap[sectionName] = append(sectionmap[sectionName], Field{
					Key:    field.Key,
					Values: field.Values,
				})
			}

			// deal with the weird quirk in the fluent-bit yaml
			// format where the name property is implicit and declared
			// as the key of the property array.
			if !sectionmap[sectionName].hasField("name") {
				sectionmap[sectionName] = append(sectionmap[sectionName], Field{
					Key: "Name",
					Values: []Value{{
						String: &sectionName,
					}},
				})
			}

			switch section.Type {
			case InputSection:
				yg.Pipeline.Inputs = append(yg.Pipeline.Inputs, sectionmap)
			case FilterSection:
				yg.Pipeline.Filters = append(yg.Pipeline.Filters, sectionmap)
			case OutputSection:
				yg.Pipeline.Outputs = append(yg.Pipeline.Outputs, sectionmap)
			case ParserSection:
				yg.Pipeline.Parsers = append(yg.Pipeline.Parsers, sectionmap)
			case CustomSection:
				yg.Customs = append(yg.Customs, sectionmap)
			}
		}
	}

	return &yg, nil
}

func (yg *yamlGrammar) dumpConfig() *Config {
	cfg := &Config{
		PluginIndex: make(map[string]int),
		Sections:    make([]ConfigSection, 0),
	}

	if yg.Service != nil {
		service := ConfigSection{
			Type:   ServiceSection,
			Fields: make([]Field, 0),
		}
		for _, field := range yg.Service {
			service.Fields = append(service.Fields, field)
		}
		cfg.Sections = append(cfg.Sections, service)
	}

	for _, fields := range yg.Pipeline.Inputs {
		pluginName := getPluginName(fields)
		pluginIndex, ok := cfg.PluginIndex[pluginName]
		if !ok {
			pluginIndex = 0
		}
		id := fmt.Sprintf("%s.%d", pluginName, pluginIndex)
		section := ConfigSection{
			Type:   InputSection,
			ID:     id,
			Fields: make([]Field, 0),
		}
		for _, fields := range fields {
			section.Fields = append(section.Fields, fields...)
		}
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range yg.Pipeline.Filters {
		pluginName := getPluginName(fields)
		pluginIndex, ok := cfg.PluginIndex[pluginName]
		if !ok {
			pluginIndex = 0
		}
		id := fmt.Sprintf("%s.%d", pluginName, pluginIndex)
		section := ConfigSection{
			Type:   FilterSection,
			ID:     id,
			Fields: make([]Field, 0),
		}
		for _, fields := range fields {
			section.Fields = append(section.Fields, fields...)
		}
		cfg.Sections = append(cfg.Sections, section)
	}
	for _, fields := range yg.Pipeline.Outputs {
		pluginName := getPluginName(fields)
		pluginIndex, ok := cfg.PluginIndex[pluginName]
		if !ok {
			pluginIndex = 0
		}
		id := fmt.Sprintf("%s.%d", pluginName, pluginIndex)
		section := ConfigSection{
			Type:   OutputSection,
			ID:     id,
			Fields: make([]Field, 0),
		}
		for _, fields := range fields {
			section.Fields = append(section.Fields, fields...)
		}
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range yg.Pipeline.Parsers {
		pluginName := getPluginName(fields)
		pluginIndex, ok := cfg.PluginIndex[pluginName]
		if !ok {
			pluginIndex = 0
		}
		id := fmt.Sprintf("%s.%d", pluginName, pluginIndex)
		section := ConfigSection{
			Type:   ParserSection,
			ID:     id,
			Fields: make([]Field, 0),
		}
		for _, fields := range fields {
			section.Fields = append(section.Fields, fields...)
		}
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range yg.Customs {
		pluginName := getPluginName(fields)
		pluginIndex, ok := cfg.PluginIndex[pluginName]
		if !ok {
			pluginIndex = 0
		}
		id := fmt.Sprintf("%s.%d", pluginName, pluginIndex)
		section := ConfigSection{
			Type:   CustomSection,
			ID:     id,
			Fields: make([]Field, 0),
		}
		for _, fields := range fields {
			section.Fields = append(section.Fields, fields...)
		}
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, include := range yg.Includes {
		section := ConfigSection{
			Type: IncludeSection,
			Fields: []Field{{
				Key: "@include",
				Values: []Value{{
					String: ptr(include),
				}},
			}},
		}
		cfg.Sections = append(cfg.Sections, section)
	}

	return cfg
}

func DumpYAML(cfg *Config) ([]byte, error) {
	if y, err := cfg.dumpYamlGrammar(); err != nil {
		return []byte(""), err
	} else {
		return yaml.Marshal(y)
	}
}

func ParseYAML(data []byte) (*Config, error) {
	var g yamlGrammar
	err := yaml.UnmarshalStrict(data, &g)
	if err != nil {
		return nil, err
	}

	return g.dumpConfig(), nil
}

func ParseJSON(data []byte) (*Config, error) {
	var g yamlGrammar
	err := json.Unmarshal(data, &g)
	if err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			fmt.Println("unmarshal type error")
		case *json.InvalidUnmarshalError:
			fmt.Println("invalid unmarshal error")
		default:
			fmt.Printf("unknown error: %T\n", err)
		}
		return nil, err
	}

	return g.dumpConfig(), nil
}

func DumpJSON(cfg *Config) ([]byte, error) {
	if j, err := cfg.dumpYamlGrammar(); err != nil {
		return []byte(""), err
	} else {
		return json.MarshalIndent(j, "", "  ")
	}
}
