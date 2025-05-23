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
	Parsers  Plugins             `json:"parsers,omitempty" yaml:"parsers,omitempty"`
}

type Pipeline struct {
	Inputs  Plugins `json:"inputs,omitempty" yaml:"inputs,omitempty"`
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
	case SectionKindFilter:
		c.Pipeline.Filters = append(c.Pipeline.Filters, makePlugin(len(c.Pipeline.Filters)))
	case SectionKindOutput:
		c.Pipeline.Outputs = append(c.Pipeline.Outputs, makePlugin(len(c.Pipeline.Outputs)))
	case SectionKindParser:
		c.Parsers = append(c.Parsers, makePlugin(len(c.Parsers)))
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

	if !c.Pipeline.Filters.Equal(target.Pipeline.Filters) {
		return false
	}

	if !c.Pipeline.Outputs.Equal(target.Pipeline.Outputs) {
		return false
	}

	if !c.Parsers.Equal(target.Parsers) {
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
	set(SectionKindFilter, c.Pipeline.Filters)
	set(SectionKindOutput, c.Pipeline.Outputs)
	set(SectionKindParser, c.Parsers)

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

	if plugin, ok := find(SectionKindFilter, c.Pipeline.Filters, id); ok {
		return plugin, true
	}

	if plugin, ok := find(SectionKindParser, c.Parsers, id); ok {
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
