package fluentbitconfig

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

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

	if !equalByNames(c.Customs, target.Customs) {
		return false
	}

	if !equalByNames(c.Pipeline.Inputs, target.Pipeline.Inputs) {
		return false
	}

	if !equalByNames(c.Pipeline.Parsers, target.Pipeline.Parsers) {
		return false
	}

	if !equalByNames(c.Pipeline.Filters, target.Pipeline.Filters) {
		return false
	}

	if !equalByNames(c.Pipeline.Outputs, target.Pipeline.Outputs) {
		return false
	}

	return true
}

func (c Config) IDs(withPrefix bool) []string {
	collect := func(kind SectionKind, byNames []ByName) []string {
		var ids []string
		for i, byName := range byNames {
			for name, props := range byName {
				if s := Name(props); s != "" {
					name = s
				}
				if withPrefix {
					ids = append(ids, fmt.Sprintf("%s:%s:%s.%d", kind, name, name, i))
				} else {
					ids = append(ids, fmt.Sprintf("%s.%d", name, i))
				}
				break
			}
		}
		return ids
	}

	var ids []string
	ids = append(ids, collect(SectionKindCustom, c.Customs)...)
	ids = append(ids, collect(SectionKindInput, c.Pipeline.Inputs)...)
	ids = append(ids, collect(SectionKindParser, c.Pipeline.Parsers)...)
	ids = append(ids, collect(SectionKindFilter, c.Pipeline.Filters)...)
	ids = append(ids, collect(SectionKindOutput, c.Pipeline.Outputs)...)
	return ids
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

func equalByNames(a, b []ByName) bool {
	return slices.EqualFunc(a, b, equalByName)
}

func equalByName(a, b ByName) bool {
	return maps.EqualFunc(a, b, func(v1, v2 property.Properties) bool {
		return v1.Equal(v2)
	})
}
