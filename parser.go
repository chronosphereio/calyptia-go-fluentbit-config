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
		{"TemplateVariable", `\{\{[^\}]+\}\}`},
		{"JsonObject", `\{[^\}]+\}`},
		{"DateTime", `\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d(\.\d+)?`},
		{"Date", `\d\d\d\d-\d\d-\d\d`},
		{"Time", `\d\d:\d\d:\d\d(\.\d+)?`},
		{"TimeFormat", `((%([aAbBhcCdeDHIjmMnprRStTUwWxXyYLz]|E[cCxXyY]|O[deHImMSUwWy]))(\:|\/| |\-|T|Z|\.|))+`},
		{"Address", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`},
		{"Unit", `[-+]?(\d*\.)?\d+([kKmMgG])`},
		{"Float", `[-+]?(\d*\.)?\d+`},
		{"Number", `[-+]?(\d*)?\d+`},
		{"Ident", `[a-zA-Z_\.\*\-0-9]+`},
		{"List", `([a-zA-Z_\.\*\-0-9\/]+ )+[a-zA-Z_\.\*\-0-9\/]+`},
		{"String", `[a-zA-Z_\.\*\-0-9\/\$\%\\\/]+`},
		{"Punct", `\[|]|[-!()+/*=,]`},
		{"Regex", `\^[^\n]+\$`},
		{"Topic", `\$[a-zA-Z0-9_\.\/\*\-]+`},
		{"comment", `#[^\n]+`},
		{"whitespace", `[ \t]+`},
		{"EOL", "[\r\n]*"},
	}
)

type ConfigGrammar struct {
	Pos     lexer.Position
	Start   string   `@EOL*`
	Entries []*Entry `@@*`
	End     string   `@EOL*`
}

type Entry struct {
	Field   *Field   `  @@`
	Section *Section `| @@`
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
	}
	return false
}

func (v *Value) Value() interface{} {
	switch {
	case v.TemplateVariable != nil:
		return *v.TemplateVariable
	case v.JsonObject != nil:
		return *v.JsonObject
	case v.String != nil:
		return *v.String
	case v.DateTime != nil:
		return *v.DateTime
	case v.Date != nil:
		return *v.Date
	case v.Time != nil:
		return *v.Time
	case v.TimeFormat != nil:
		return *v.TimeFormat
	case v.Topic != nil:
		return *v.Topic
	case v.Regex != nil:
		return *v.Regex
	case v.Bool != nil:
		return *v.Bool
	case v.Number != nil:
		return *v.Number
	case v.Float != nil:
		return *v.Float
	}
	return nil
}

func (v *Value) ToString() string {
	val := v.Value()
	switch t := val.(type) {
	case string:
		return t
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		return fmt.Sprintf("%0.6f", t)
	case bool:
		if t == true {
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
)

var ConfigSectionTypes []string = []string{
	"NONE",
	"SERVICE",
	"INPUT",
	"FILTER",
	"OUTPUT",
	"CUSTOM",
	"PARSER",
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

func stringPtr(s string) *string {
	return &s
}

func numberPtr(n int64) *int64 {
	return &n
}

func floatPtr(f float64) *float64 {
	return &f
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

func (c *Config) loadSectionsFromGrammar(grammar *ConfigGrammar) error {
	for _, entry := range grammar.Entries {
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
	}
	return nil
}

func NewFromBytes(data []byte, skipProperties ...string) (*Config, error) {
	if len(skipProperties) > 0 {
		fmt.Println("deprecated: we no longer support skipping properties")
	}
	return ParseINI(data)
}

func (value *Value) dumpValueINI(ini *bytes.Buffer) error {
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
	case value.String != nil:
		ini.Write([]byte(*value.String))
	case value.DateTime != nil:
		ini.Write([]byte(*value.DateTime))
	case value.Date != nil:
		ini.Write([]byte(*value.Date))
	case value.Time != nil:
		ini.Write([]byte(*value.Time))
	case value.TimeFormat != nil:
		ini.Write([]byte(*value.TimeFormat))
	case value.Topic != nil:
		ini.Write([]byte(fmt.Sprintf("$%s", *value.Topic)))
	case value.Regex != nil:
		ini.Write([]byte(*value.Regex))
	case value.Bool != nil:
		if *value.Bool {
			ini.Write([]byte("on"))
		} else {
			ini.Write([]byte("off"))
		}
	case value.Number != nil:
		ini.Write([]byte(fmt.Sprintf("%d", int(*value.Number))))
	case value.Float != nil:
		ini.Write([]byte(fmt.Sprintf("%0.6f", *value.Float)))
	case value.List != nil:
		//for _, v := range value.List {
		//if err := v.dumpValueINI(ini); err != nil {
		//	return err
		//}
		//}
	}
	return nil
}

func (values Values) dumpValueINI(ini *bytes.Buffer) error {
	for _, value := range values {
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

func (c *Config) DumpINI() ([]byte, error) {
	// type Config struct {
	// Inputs      map[string][]Field
	// Filters     map[string][]Field
	// Customs     map[string][]Field
	// Outputs     map[string][]Field
	// Parsers     map[string][]Field

	ini := bytes.NewBuffer([]byte(""))

	for _, section := range c.Sections {
		ini.Write([]byte(fmt.Sprintf("[%s]\n",
			ConfigSectionTypes[section.Type])))

		for _, field := range section.Fields {
			if err := field.dumpFieldINI(ini); err != nil {
				return []byte(""), err
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

func (fs *Fields) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	kv := make(map[string]interface{})
	if err := unmarshal(&kv); err != nil {
		return err
	}
	fields := make([]Field, 0)
	for k, v := range kv {
		switch t := v.(type) {
		case string:
			fields = append(fields, Field{
				Key: k,
				Values: []Value{{
					String: &t,
				}},
			})
		case int:
			i := int64(v.(int))
			fields = append(fields, Field{
				Key: k,
				Values: []Value{{
					Number: &i,
				}},
			})
		case int64:
			fields = append(fields, Field{
				Key: k,
				Values: []Value{{
					Number: &t,
				}},
			})
		case float64:
			fields = append(fields, Field{
				Key: k,
				Values: []Value{{
					Float: &t,
				}},
			})
		case bool:
			fields = append(fields, Field{
				Key: k,
				Values: []Value{{
					Bool: &t,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v: %+v", t, v)
		}
	}
	*fs = fields

	return nil
}

func (fs Fields) MarshalYAML() (interface{}, error) {
	fmap := make(map[string]interface{})
	for _, field := range fs {
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

func (fs Fields) MarshalJSON() ([]byte, error) {
	bfields := bytes.NewBuffer([]byte(""))
	bfields.Write([]byte("{"))

	for fidx, field := range fs {
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

		if fidx < len(fs)-1 {
			bfields.Write([]byte(", "))
		}
	}

	bfields.Write([]byte("}"))
	return bfields.Bytes(), nil
}

func (fs *Fields) UnmarshalJSON(b []byte) error {
	fields := make(Fields, 0)

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
			fields = append(fields, Field{
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
				fields = append(fields, Field{
					Key: k.(string),
					Values: []Value{{
						Float: &f,
					}},
				})
			} else {
				fields = append(fields, Field{
					Key: k.(string),
					Values: []Value{{
						Number: &i,
					}},
				})
			}
		case bool:
			fields = append(fields, Field{
				Key: k.(string),
				Values: []Value{{
					Bool: &t,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v", v)
		}
	}

	*fs = fields
	return nil
}

type yamlGrammarPipeline struct {
	Inputs  []map[string]Fields `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Filters []map[string]Fields `yaml:"filters,omitempty" json:"filters,omitempty"`
	Parsers []map[string]Fields `yaml:"parsers,omitempty" json:"parsers,omitempty"`
	Outputs []map[string]Fields `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

type yamlGrammar struct {
	Service  map[string]interface{} `yaml:"service,omitempty" json:"service,omitempty"`
	Customs  []map[string]Fields    `yaml:"customs,omitempty" json:"customs,omitempty"`
	Pipeline yamlGrammarPipeline    `yaml:"pipeline" json:"pipeline"`
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

func (cfg *Config) dumpYamlGrammar() *yamlGrammar {
	yg := yamlGrammar{
		Service: make(map[string]interface{}),
		Customs: make([]map[string]Fields, 0),
		Pipeline: yamlGrammarPipeline{
			Inputs:  make([]map[string]Fields, 0),
			Filters: make([]map[string]Fields, 0),
			Outputs: make([]map[string]Fields, 0),
			Parsers: make([]map[string]Fields, 0),
		},
	}

	for _, section := range cfg.Sections {
		if section.Type == ServiceSection {
			for _, field := range section.Fields {
				if len(field.Values) == 1 {
					switch true {
					case field.Values[0].String != nil:
						yg.Service[field.Key] = *field.Values[0].String
					case field.Values[0].DateTime != nil:
						yg.Service[field.Key] = *field.Values[0].DateTime
					case field.Values[0].Date != nil:
						yg.Service[field.Key] = *field.Values[0].Date
					case field.Values[0].Time != nil:
						yg.Service[field.Key] = *field.Values[0].Time
					case field.Values[0].Bool != nil:
						yg.Service[field.Key] = *field.Values[0].Bool
					case field.Values[0].Number != nil:
						yg.Service[field.Key] = *field.Values[0].Number
					case field.Values[0].Float != nil:
						yg.Service[field.Key] = *field.Values[0].Float
					}
				} else {
					yg.Service[field.Key] = field.Values.ToString()
				}
			}
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

	return &yg
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
		for k, v := range yg.Service {
			value := make([]Value, 1)
			switch v.(type) {
			case string:
				value[0].String = stringPtr(v.(string))
			case int:
				value[0].Number = numberPtr(int64(v.(int)))
			case int64:
				value[0].Number = numberPtr(v.(int64))
			case float64:
				value[0].Float = floatPtr(v.(float64))
			}
			service.Fields = append(service.Fields, Field{
				Key:    k,
				Values: value,
			})
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

	return cfg
}

func DumpYAML(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg.dumpYamlGrammar())
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
		case *json.UnmarshalFieldError:
			fmt.Println("unmarshal field error")
		case *json.InvalidUTF8Error:
			fmt.Println("syntax error")
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
	return json.MarshalIndent(cfg.dumpYamlGrammar(), "", "  ")
}
