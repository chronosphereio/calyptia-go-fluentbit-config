package fluentbitconfig

import (
	"fmt"
	"strings"

	"github.com/calyptia/go-fluentbit-config/property"
)

type Config struct {
	Env      property.Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Includes []string            `json:"includes,omitempty" yaml:"includes,omitempty"`
	Service  property.Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  []ByName            `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline            `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
}

type Pipeline struct {
	Inputs  []ByName `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Parsers []ByName `json:"parsers,omitempty" yaml:"parsers,omitempty"`
	Filters []ByName `json:"filters,omitempty" yaml:"filters,omitempty"`
	Outputs []ByName `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type ByName map[string]property.Properties

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

	byName := ByName{
		name: props,
	}

	switch kind {
	case SectionKindCustom:
		c.Customs = append(c.Customs, byName)
	case SectionKindInput:
		c.Pipeline.Inputs = append(c.Pipeline.Inputs, byName)
	case SectionKindParser:
		c.Pipeline.Parsers = append(c.Pipeline.Parsers, byName)
	case SectionKindFilter:
		c.Pipeline.Filters = append(c.Pipeline.Filters, byName)
	case SectionKindOutput:
		c.Pipeline.Outputs = append(c.Pipeline.Outputs, byName)
	}
}

// Name from properties.
func Name(props property.Properties) string {
	nameVal, ok := props.Get("name")
	if !ok {
		return ""
	}

	if name, ok := nameVal.(string); ok {
		return strings.TrimSpace(strings.ToValidUTF8(name, ""))
	}

	return strings.TrimSpace(strings.ToValidUTF8(fmt.Sprintf("%v", nameVal), ""))
}
