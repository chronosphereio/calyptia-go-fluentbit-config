package fluentbitconfig

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

type Config struct {
	Env      property.Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Includes []string            `json:"includes,omitempty" yaml:"includes,omitempty"`
	Service  property.Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  Plugins             `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline            `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
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

// Name from properties.
func Name(props property.Properties) string {
	nameVal, ok := props.Get("name")
	if !ok {
		return ""
	}

	name := stringFromAny(nameVal)
	return strings.ToLower(strings.TrimSpace(strings.ToValidUTF8(name, "")))
}
