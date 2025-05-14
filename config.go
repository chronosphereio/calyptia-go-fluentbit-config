package fluentbitconfig

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

type Config struct {
	Env      property.Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Includes []string            `json:"includes,omitempty" yaml:"includes,omitempty"`
	Service  property.Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  Plugins             `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline            `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	Parsers  []Parser            `json:"parsers,omitempty" yaml:"parsers,omitempty"`
}

type ConfigBool bool

func (cb *ConfigBool) UnmarshalYAML(value *yaml.Node) error {
	var configBoolStr string

	if err := value.Decode(&configBoolStr); err != nil {
		return err
	}

	switch ParserFormat(configBoolStr) {
	case "on":
		fallthrough
	case "On":
		fallthrough
	case "ON":
		fallthrough
	case "true":
		*cb = true
	case "off":
		fallthrough
	case "Off":
		fallthrough
	case "OFF":
		fallthrough
	case "false":
		*cb = false
	default:
		return fmt.Errorf("unknown boolean value: %s", string(configBoolStr))
	}
	return nil
}

type ParserFormat string

const (
	ParserFormatJson   ParserFormat = "json"
	ParserFormatRegex  ParserFormat = "regex"
	ParserFormatLtsv   ParserFormat = "ltsv"
	ParserFormatLogfmt ParserFormat = "logfmt"
)

func (pf *ParserFormat) UnmarshalYAML(value *yaml.Node) error {
	var parserFormatString string

	if err := value.Decode(&parserFormatString); err != nil {
		return err
	}

	switch ParserFormat(parserFormatString) {
	case ParserFormatJson:
		*pf = ParserFormatJson
	case ParserFormatRegex:
		*pf = ParserFormatRegex
	case ParserFormatLtsv:
		*pf = ParserFormatLtsv
	case ParserFormatLogfmt:
		*pf = ParserFormatLogfmt
	default:
		return fmt.Errorf("unknown parser type: %s", string(parserFormatString))
	}
	return nil
}

type Parser struct {
	Name               string       `json:"name" yaml:"name"`
	Format             ParserFormat `json:"format" yaml:"format"`
	Regex              string       `json:"regex,omitempty" yaml:"regex,omitempty"`
	TimeKey            string       `json:"time_key,omitempty" yaml:"time_key,omitempty"`
	TimeFormat         string       `json:"time_format,omitempty" yaml:"time_format,omitempty"`
	TimeOffset         string       `json:"time_offset,omitempty" yaml:"time_offset,omitempty"`
	TimeKeep           ConfigBool   `json:"time_keep,omitempty" yaml:"time_keep,omitempty"`
	TimeSystemTimezone ConfigBool   `json:"time_system_timezone,omitempty" yaml:"time_system_timezone,omitempty"`
	Types              string       `json:"types,omitempty" yaml:"types,omitempty"`
	SkipEmptyValues    ConfigBool   `json:"skip_empty_values,omitempty" yaml:"skip_empty_values,omitempty"`
	TimeStrict         ConfigBool   `json:"time_strict,omitempty" yaml:"time_strict,omitempty"`
}

type Pipeline struct {
	Inputs  Plugins `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Parsers Plugins `json:"parsers,omitempty" yaml:"parsers,omitempty"`
	Filters Plugins `json:"filters,omitempty" yaml:"filters,omitempty"`
	Outputs Plugins `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

func (c *Config) SetEnv(key string, value any) {
	if c.Env == nil {
		c.Env = property.Properties{}
	}
	c.Env.Set(key, value)
}

func (c *Config) Include(path string) {
	c.Includes = append(c.Includes, path)
}

func (c *Config) AddSection(kind SectionKind, props property.Properties) {
	if kind == SectionKindService {
		if c.Service == nil {
			c.Service = property.Properties{}
		}
		for _, p := range props {
			c.Service.Set(p.Key, p.Value)
		}
		return
	}

	name := Name(props)
	if name == "" {
		return
	}

	makePlugin := func(i int) Plugin {
		return Plugin{
			ID:         fmt.Sprintf("%s.%d", name, i),
			Name:       name,
			Properties: props,
		}
	}

	switch kind {
	case SectionKindCustom:
		c.Customs = append(c.Customs, makePlugin(len(c.Customs)))
	case SectionKindInput:
		c.Pipeline.Inputs = append(c.Pipeline.Inputs, makePlugin(len(c.Pipeline.Inputs)))
	case SectionKindParser:
		c.Pipeline.Parsers = append(c.Pipeline.Parsers, makePlugin(len(c.Pipeline.Parsers)))
	case SectionKindFilter:
		c.Pipeline.Filters = append(c.Pipeline.Filters, makePlugin(len(c.Pipeline.Filters)))
	case SectionKindOutput:
		c.Pipeline.Outputs = append(c.Pipeline.Outputs, makePlugin(len(c.Pipeline.Outputs)))
	}
}

func (c Config) Equal(target Config) bool {
	if !c.Env.Equal(target.Env) {
		return false
	}

	if !slices.Equal(c.Includes, target.Includes) {
		return false
	}

	if !c.Service.Equal(target.Service) {
		return false
	}

	if !c.Customs.Equal(target.Customs) {
		return false
	}

	if !c.Pipeline.Inputs.Equal(target.Pipeline.Inputs) {
		return false
	}

	if !c.Pipeline.Parsers.Equal(target.Pipeline.Parsers) {
		return false
	}

	if !c.Pipeline.Filters.Equal(target.Pipeline.Filters) {
		return false
	}

	if !c.Pipeline.Outputs.Equal(target.Pipeline.Outputs) {
		return false
	}

	return true
}

// IDs namespaced with the section kind and name.
// For example: input:tail:tail.0
func (c Config) IDs(namespaced bool) []string {
	var ids []string
	set := func(kind SectionKind, plugins Plugins) {
		for _, plugin := range plugins {
			if namespaced {
				ids = append(ids, fmt.Sprintf("%s:%s:%s", kind, plugin.Name, plugin.ID))
			} else {
				ids = append(ids, plugin.ID)
			}
		}
	}

	set(SectionKindCustom, c.Customs)
	set(SectionKindInput, c.Pipeline.Inputs)
	set(SectionKindParser, c.Pipeline.Parsers)
	set(SectionKindFilter, c.Pipeline.Filters)
	set(SectionKindOutput, c.Pipeline.Outputs)

	return ids
}

// FindByID were the id should be namespaced with the section kind and name.
// For example: input:tail:tail.0
func (c Config) FindByID(id string) (Plugin, bool) {
	find := func(kind SectionKind, plugins Plugins, id string) (Plugin, bool) {
		for _, plugin := range plugins {
			if fmt.Sprintf("%s:%s:%s", kind, plugin.Name, plugin.ID) == id {
				return plugin, true
			}
		}
		return Plugin{}, false
	}

	if plugin, ok := find(SectionKindCustom, c.Customs, id); ok {
		return plugin, true
	}

	if plugin, ok := find(SectionKindInput, c.Pipeline.Inputs, id); ok {
		return plugin, true
	}

	if plugin, ok := find(SectionKindParser, c.Pipeline.Parsers, id); ok {
		return plugin, true
	}

	if plugin, ok := find(SectionKindFilter, c.Pipeline.Filters, id); ok {
		return plugin, true
	}

	return Plugin{}, false
}

// Name from properties.
func Name(props property.Properties) string {
	nameVal, ok := props.Get("name")
	if !ok {
		return ""
	}

	name := stringFromAny(nameVal)
	return strings.ToLower(strings.TrimSpace(strings.ToValidUTF8(name, "")))
}
