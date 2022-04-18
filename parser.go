package fluentbit_config

import (
	"bytes"
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

type Config struct {
	Service     []Field
	Inputs      map[string][]Field
	Filters     map[string][]Field
	Customs     map[string][]Field
	Outputs     map[string][]Field
	Parsers     map[string][]Field
	inputIndex  int
	filterIndex int
	outputIndex int
	customIndex int
	parserIndex int
}

func addServiceFields(e *Entry, f *[]Field) {
	for _, field := range e.Section.Fields {
		*f = append(*f, *field)
	}
}

func addFields(e *Entry, index int, m *map[string][]Field) {
	if *m == nil {
		*m = map[string][]Field{}
	}
	q := *m
	var name string
	for _, field := range e.Section.Fields {
		if strings.ToLower(field.Key) == "name" {
			name = fmt.Sprintf("%s.%d", *field.Values[0].String, index)
		}
	}

	for _, field := range e.Section.Fields {
		q[name] = append(q[name], *field)
	}
}

func (c *Config) loadSectionsFromGrammar(grammar *ConfigGrammar) error {
	for _, entry := range grammar.Entries {
		switch entry.Section.Name {
		case "SERVICE":
			{
				addServiceFields(entry, &c.Service)
			}
		case "INPUT":
			{
				addFields(entry, c.inputIndex, &c.Inputs)
				c.inputIndex++
			}
		case "FILTER":
			{
				addFields(entry, c.filterIndex, &c.Filters)
				c.filterIndex++
			}
		case "OUTPUT":
			{
				addFields(entry, c.outputIndex, &c.Outputs)
				c.outputIndex++
			}
		case "CUSTOM":
			{
				addFields(entry, c.customIndex, &c.Customs)
				c.customIndex++
			}
		case "PARSER":
			{
				addFields(entry, c.parserIndex, &c.Parsers)
				c.parserIndex++
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

	for _, fields := range c.Inputs {
		ini.Write([]byte("[INPUT]\n"))

		for _, field := range fields {
			if err := field.dumpFieldINI(ini); err != nil {
				return []byte(""), err
			}
		}
	}

	for _, fields := range c.Filters {
		ini.Write([]byte("[FILTER]\n"))

		for _, field := range fields {
			if err := field.dumpFieldINI(ini); err != nil {
				return []byte(""), err
			}
		}
	}

	for _, fields := range c.Customs {
		ini.Write([]byte("[CUSTOM]\n"))

		for _, field := range fields {
			if err := field.dumpFieldINI(ini); err != nil {
				return []byte(""), err
			}
		}
	}

	for _, fields := range c.Outputs {
		ini.Write([]byte("[OUTPUT]\n"))

		for _, field := range fields {
			if err := field.dumpFieldINI(ini); err != nil {
				return []byte(""), err
			}
		}
	}

	for _, fields := range c.Parsers {
		ini.Write([]byte("[PARSER]\n"))

		for _, field := range fields {
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
	var config = Config{}

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

type yamlGrammarPipeline struct {
	Inputs  []map[string]Fields `yaml:"inputs"`
	Filters []map[string]Fields `yaml:"filters,omitempty"`
	Parsers []map[string]Fields `yaml:"parsers,omitempty"`
	Outputs []map[string]Fields `yaml:"outputs"`
	Customs []map[string]Fields `yaml:"customs,omitempty"`
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

type yamlGrammar struct {
	Service  map[string]interface{} `yaml:"service"`
	Pipeline yamlGrammarPipeline    `yaml:"pipeline"`
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

	for _, v := range cfg.Service {
		switch true {
		case v.Values[0].String != nil:
			yg.Service[v.Key] = *v.Values[0].String
		case v.Values[0].DateTime != nil:
			yg.Service[v.Key] = *v.Values[0].DateTime
		case v.Values[0].Date != nil:
			yg.Service[v.Key] = *v.Values[0].Date
		case v.Values[0].Time != nil:
			yg.Service[v.Key] = *v.Values[0].Time
		case v.Values[0].Bool != nil:
			yg.Service[v.Key] = *v.Values[0].Bool
		case v.Values[0].Number != nil:
			yg.Service[v.Key] = *v.Values[0].Number
		case v.Values[0].Float != nil:
			yg.Service[v.Key] = *v.Values[0].Float
		}
	}

	/************************
	type Config struct {
		Service     []Field
		Inputs      map[string][]Field
		Filters     map[string][]Field
		Customs     map[string][]Field
		Outputs     map[string][]Field
		Parsers     map[string][]Field
		inputIndex  int
		filterIndex int
		outputIndex int
		customIndex int
		parserIndex int
	}
	**********************/
	for _, fields := range cfg.Inputs {
		sectionName := getPluginNameParameter(fields)
		section := make(map[string]Fields)
		section[sectionName] = make(Fields, 0)

		for _, field := range fields {
			section[sectionName] = append(section[sectionName], Field{
				Key:    field.Key,
				Values: field.Values,
			})
		}
		yg.Pipeline.Inputs = append(yg.Pipeline.Inputs, section)
	}

	for _, fields := range cfg.Filters {
		sectionName := getPluginNameParameter(fields)
		section := make(map[string]Fields)
		section[sectionName] = make(Fields, 0)

		for _, field := range fields {
			section[sectionName] = append(section[sectionName], Field{
				Key:    field.Key,
				Values: field.Values,
			})
		}
		yg.Pipeline.Filters = append(yg.Pipeline.Filters, section)
	}

	for _, fields := range cfg.Outputs {
		sectionName := getPluginNameParameter(fields)
		section := make(map[string]Fields)
		section[sectionName] = make(Fields, 0)

		for _, field := range fields {
			section[sectionName] = append(section[sectionName], Field{
				Key:    field.Key,
				Values: field.Values,
			})
		}
		yg.Pipeline.Outputs = append(yg.Pipeline.Outputs, section)
	}

	for _, fields := range cfg.Parsers {
		sectionName := getPluginNameParameter(fields)
		section := make(map[string]Fields)
		section[sectionName] = make(Fields, 0)

		for _, field := range fields {
			section[sectionName] = append(section[sectionName], Field{
				Key:    field.Key,
				Values: field.Values,
			})
		}
		yg.Pipeline.Parsers = append(yg.Pipeline.Parsers, section)
	}

	for _, fields := range cfg.Customs {
		sectionName := getPluginNameParameter(fields)
		section := make(map[string]Fields)
		section[sectionName] = make(Fields, 0)

		for _, field := range fields {
			section[sectionName] = append(section[sectionName], Field{
				Key:    field.Key,
				Values: field.Values,
			})
		}
		yg.Pipeline.Customs = append(yg.Pipeline.Customs, section)
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
		Service: make([]Field, 0),
		Inputs:  make(map[string][]Field),
		Filters: make(map[string][]Field),
		Outputs: make(map[string][]Field),
		Parsers: make(map[string][]Field),
		Customs: make(map[string][]Field),
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
		cfg.Service = append(cfg.Service, Field{
			Key:    k,
			Values: []*Value{&value},
		})
	}

	for idx, fields := range g.Pipeline.Inputs {
		pluginName := getPluginName(fields)
		section := fmt.Sprintf("%s.%d", pluginName, idx)
		cfg.Inputs[section] = make([]Field, 0)
		for _, fields := range fields {
			cfg.Inputs[section] = append(cfg.Inputs[section], fields...)
		}
		cfg.Inputs[section] = append(cfg.Inputs[section], Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
	}

	for _, fields := range g.Pipeline.Filters {
		pluginName := getPluginName(fields)
		section := fmt.Sprintf("%s.%d", pluginName, cfg.filterIndex)
		cfg.Filters[section] = make([]Field, 0)
		for _, fields := range fields {
			cfg.Filters[section] = append(cfg.Filters[section], fields...)
		}
		cfg.Filters[section] = append(cfg.Filters[section], Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.filterIndex++
	}

	for _, fields := range g.Pipeline.Outputs {
		pluginName := getPluginName(fields)
		section := fmt.Sprintf("%s.%d", pluginName, cfg.outputIndex)
		cfg.Outputs[section] = make([]Field, 0)
		for _, fields := range fields {
			cfg.Outputs[section] = append(cfg.Outputs[section], fields...)
		}
		cfg.Outputs[section] = append(cfg.Outputs[section], Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.outputIndex++
	}

	for _, fields := range g.Pipeline.Customs {
		pluginName := getPluginName(fields)
		section := fmt.Sprintf("%s.%d", pluginName, cfg.customIndex)
		cfg.Customs[section] = make([]Field, 0)
		for _, fields := range fields {
			cfg.Customs[section] = append(cfg.Customs[section], fields...)
		}
		cfg.Customs[section] = append(cfg.Customs[section], Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.customIndex++
	}

	for _, fields := range g.Pipeline.Parsers {
		pluginName := getPluginName(fields)
		section := fmt.Sprintf("%s.%d", pluginName, cfg.parserIndex)
		cfg.Parsers[section] = make([]Field, 0)
		for _, fields := range fields {
			cfg.Parsers[section] = append(cfg.Parsers[section], fields...)
		}
		cfg.Parsers[section] = append(cfg.Parsers[section], Field{
			Key:    "name",
			Values: []*Value{{String: stringPtr(pluginName)}},
		})
		cfg.parserIndex++
	}

	return cfg, nil
}

func ParseJSON(data []byte) (*Config, error) {
	return nil, fmt.Errorf("WIP, unimplemented")
}
