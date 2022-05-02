package fluentbit_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gopkg.in/yaml.v2"
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
		{"whitespace", `[\ \t]+`, nil},
		{"EOL", "[\r\n]+", nil},
	}
)

type ConfigGrammar struct {
	Pos     lexer.Position
	Start   string   ` @EOL*`
	Entries []*Entry `@@*`
	End     string   ` @EOL*`
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
	Key    string   `@Ident`
	Values []*Value `@@+`
	End    string   ` @EOL*`
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

func numberPtr(n float64) *float64 {
	return &n
}

func (c *Config) addSection(sectype ConfigSectionType, e *Entry) {
	var plugin string
	section := ConfigSection{Type: sectype}

	for _, field := range e.Section.Fields {
		if strings.ToLower(field.Key) == "name" {
			plugin = *field.Values[0].String
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
		ini.Write([]byte(fmt.Sprintf("%s", *value.String)))
	case value.DateTime != nil:
		ini.Write([]byte(fmt.Sprintf("%s", *value.DateTime)))
	case value.Date != nil:
		ini.Write([]byte(fmt.Sprintf("%s", *value.Date)))
	case value.Time != nil:
		ini.Write([]byte(fmt.Sprintf("%s", *value.Time)))
	case value.TimeFormat != nil:
		ini.Write([]byte(fmt.Sprintf("%s", *value.TimeFormat)))
	case value.Topic != nil:
		ini.Write([]byte(fmt.Sprintf("$%s", *value.Topic)))
	case value.Regex != nil:
		ini.Write([]byte(fmt.Sprintf("%s", *value.Regex)))
	case value.Bool != nil:
		if *value.Bool {
			ini.Write([]byte("on"))
		} else {
			ini.Write([]byte("off"))
		}
	case value.Float != nil:
		ini.Write([]byte(fmt.Sprintf("%f", *value.Float)))
	case value.Number != nil:
		ini.Write([]byte(fmt.Sprintf("%f", *value.Number)))
	case value.List != nil:
		for _, v := range value.List {
			if err := v.dumpValueINI(ini); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Field) dumpFieldINI(ini *bytes.Buffer) error {
	ini.Write([]byte(fmt.Sprintf("%s ", f.Key)))
	for idx, value := range f.Values {
		if err := value.dumpValueINI(ini); err != nil {
			return err
		}
		// add whitespace between fields
		if (idx + 1) < len(f.Values) {
			ini.Write([]byte(" "))
		}
	}
	ini.Write([]byte(f.End))
	return nil
}

func (c *Config) DumpINI() ([]byte, error) {
	// type Config struct {
	// Inputs      map[string][]Field
	// Filters     map[string][]Field
	// Customs     map[string][]Field
	// Outputs     map[string][]Field
	// Parsers     map[string][]Field

	ini := bytes.NewBuffer([]byte("\n"))

	for _, section := range c.Sections {
		ini.Write([]byte(fmt.Sprintf("[%s]\n", ConfigSectionTypes[section.Type])))

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
		if k == "name" {
			continue
		}
		switch v.(type) {
		case string:
			str := v.(string)
			fields = append(fields, Field{
				Key: k,
				Values: []*Value{{
					String: &str,
				}},
			})
		case int:
			i := v.(int)
			f := float64(i)
			fields = append(fields, Field{
				Key: k,
				Values: []*Value{{
					Number: &f,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v", v)
		}
	}
	*fs = fields

	return nil
}

func (fs Fields) MarshalYAML() (interface{}, error) {
	fmap := make(map[string]interface{})
	for _, field := range fs {
		for _, value := range field.Values {
			switch true {
			case value.String != nil:
				fmap[field.Key] = *value.String
			case value.DateTime != nil:
				fmap[field.Key] = *value.DateTime
			case value.Date != nil:
				fmap[field.Key] = *value.Date
			case value.Time != nil:
				fmap[field.Key] = *value.Time
			case value.Bool != nil:
				fmap[field.Key] = *value.Bool
			case value.Number != nil:
				fmap[field.Key] = *value.Number
			case value.Float != nil:
				fmap[field.Key] = *value.Float
			default:
				return nil, fmt.Errorf("unknown type: %+v", field)
			}
		}
	}
	return fmap, nil
}

func (fs *Fields) UnmarshalJSON(raw []byte) error {
	panic("FOO FIGHTORZ")
	kv := make(map[string]interface{})
	if err := json.Unmarshal(raw, &kv); err != nil {
		return err
	}
	fields := make([]Field, 0)
	for k, v := range kv {
		if k == "name" {
			continue
		}
		switch v.(type) {
		case string:
			str := v.(string)
			fields = append(fields, Field{
				Key: k,
				Values: []*Value{{
					String: &str,
				}},
			})
		case int:
			i := v.(int)
			f := float64(i)
			fields = append(fields, Field{
				Key: k,
				Values: []*Value{{
					Number: &f,
				}},
			})
		default:
			return fmt.Errorf("unknown type: %+v", v)
		}
	}
	*fs = fields

	return nil
}

func (fs Fields) MarshalJSON() ([]byte, error) {
	fmap := make(map[string]interface{})
	for _, field := range fs {
		for _, value := range field.Values {
			switch true {
			case value.String != nil:
				fmap[field.Key] = *value.String
			case value.DateTime != nil:
				fmap[field.Key] = *value.DateTime
			case value.Date != nil:
				fmap[field.Key] = *value.Date
			case value.Time != nil:
				fmap[field.Key] = *value.Time
			case value.Bool != nil:
				fmap[field.Key] = *value.Bool
			case value.Number != nil:
				fmap[field.Key] = *value.Number
			case value.Float != nil:
				fmap[field.Key] = *value.Float
			default:
				return nil, fmt.Errorf("unknown type: %+v", field)
			}
		}
	}
	return json.Marshal(fmap)
}

type yamlGrammarPipeline struct {
	Inputs  []map[string]Fields `yaml:"inputs" json:"inputs"`
	Filters []map[string]Fields `yaml:"filters,omitempty" json:"filters,omitempty"`
	Parsers []map[string]Fields `yaml:"parsers,omitempty" json:"parsers,omitempty"`
	Outputs []map[string]Fields `yaml:"outputs" json:"outputs"`
	Customs []map[string]Fields `yaml:"customs,omitempty" json:"customs,omitempty"`
}

type yamlGrammar struct {
	Service  map[string]interface{} `yaml:"service" json:"service"`
	Pipeline yamlGrammarPipeline    `yaml:"pipeline" json:"pipeline"`
}

func getPluginName(fields map[string]Fields) string {
	rkey := ""
	for key, _ := range fields {
		rkey = key
	}
	return rkey
}

func getPluginNameParameter(fields Fields) string {
	for _, field := range fields {
		if field.Key == "name" {
			return *field.Values[0].String
		}
	}
	return ""
}

func DumpYAML(cfg *Config) ([]byte, error) {
	yg := yamlGrammar{
		Service: make(map[string]interface{}),
		Pipeline: yamlGrammarPipeline{
			Inputs:  make([]map[string]Fields, 0),
			Filters: make([]map[string]Fields, 0),
			Outputs: make([]map[string]Fields, 0),
			Parsers: make([]map[string]Fields, 0),
			Customs: make([]map[string]Fields, 0),
		},
	}

	for _, section := range cfg.Sections {
		if section.Type == ServiceSection {
			for _, field := range section.Fields {
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
			}
		} else {
			sectionName := getPluginNameParameter(section.Fields)
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
				yg.Pipeline.Customs = append(yg.Pipeline.Customs, sectionmap)
			}
		}
	}

	return yaml.Marshal(&yg)
}

func ParseYAML(data []byte) (*Config, error) {
	var g yamlGrammar
	err := yaml.UnmarshalStrict(data, &g)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		PluginIndex: make(map[string]int),
		Sections:    make([]ConfigSection, 0),
	}

	if g.Service != nil {
		service := ConfigSection{
			Type:   ServiceSection,
			Fields: make([]Field, 0),
		}
		for k, v := range g.Service {
			value := Value{}
			switch v.(type) {
			case string:
				value.String = stringPtr(v.(string))
			case int:
				value.Number = numberPtr(float64(v.(int)))
			case float64:
				value.Float = numberPtr(v.(float64))
			}
			service.Fields = append(service.Fields, Field{
				Key:    k,
				Values: []*Value{&value},
			})
		}
		cfg.Sections = append(cfg.Sections, service)
	}

	for _, fields := range g.Pipeline.Inputs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Filters {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}
	for _, fields := range g.Pipeline.Outputs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Parsers {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Customs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	return cfg, nil
}

func ParseJSON(data []byte) (*Config, error) {
	var g yamlGrammar
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		PluginIndex: make(map[string]int),
		Sections:    make([]ConfigSection, 0),
	}

	if g.Service != nil {
		service := ConfigSection{
			Type:   ServiceSection,
			Fields: make([]Field, 0),
		}
		for k, v := range g.Service {
			value := Value{}
			switch v.(type) {
			case string:
				value.String = stringPtr(v.(string))
			case int:
				value.Number = numberPtr(float64(v.(int)))
			case float64:
				value.Float = numberPtr(v.(float64))
			}
			service.Fields = append(service.Fields, Field{
				Key:    k,
				Values: []*Value{&value},
			})
		}
		cfg.Sections = append(cfg.Sections, service)
	}

	for _, fields := range g.Pipeline.Inputs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Filters {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}
	for _, fields := range g.Pipeline.Outputs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Parsers {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	for _, fields := range g.Pipeline.Customs {
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
		section.Fields = append(section.Fields, Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.Sections = append(cfg.Sections, section)
	}

	return cfg, nil
}

func DumpJSON(cfg *Config) ([]byte, error) {
	yg := yamlGrammar{
		Service: make(map[string]interface{}),
		Pipeline: yamlGrammarPipeline{
			Inputs:  make([]map[string]Fields, 0),
			Filters: make([]map[string]Fields, 0),
			Outputs: make([]map[string]Fields, 0),
			Parsers: make([]map[string]Fields, 0),
			Customs: make([]map[string]Fields, 0),
		},
	}

	for _, section := range cfg.Sections {
		if section.Type == ServiceSection {
			for _, field := range section.Fields {
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
			}
		} else {
			sectionName := getPluginNameParameter(section.Fields)
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
				yg.Pipeline.Customs = append(yg.Pipeline.Customs, sectionmap)
			}
		}
	}

	return json.MarshalIndent(&yg, "", "  ")
}
