package fluentbitconfig

import (
	"encoding/json"
	"fmt"

	"github.com/calyptia/go-fluentbit-config/v2/property"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type Plugins []Plugin

func (plugins Plugins) IDs() []string {
	var ids []string
	for _, plugin := range plugins {
		ids = append(ids, plugin.ID)
	}
	return ids
}

func (c Config) IDs() []string {
	var ids []string
	set := func(kind SectionKind, plugins Plugins) {
		for _, plugin := range plugins {
			ids = append(ids, fmt.Sprintf("%s:%s:%s", kind, plugin.Name, plugin.ID))
		}
	}

	set(SectionKindCustom, c.Customs)
	set(SectionKindInput, c.Pipeline.Inputs)
	set(SectionKindParser, c.Pipeline.Parsers)
	set(SectionKindFilter, c.Pipeline.Filters)
	set(SectionKindOutput, c.Pipeline.Outputs)

	return ids
}

func (plugins *Plugins) UnmarshalJSON(data []byte) error {
	var dest []Plugin
	err := json.Unmarshal(data, &dest)
	if err != nil {
		return err
	}

	*plugins = dest

	for i, plugin := range *plugins {
		plugin.ID = fmt.Sprintf("%s.%d", plugin.Name, i)
		(*plugins)[i] = plugin
	}

	return nil
}

func (plugins *Plugins) UnmarshalYAML(node *yaml.Node) error {
	var dest []Plugin
	err := node.Decode(&dest)
	if err != nil {
		return err
	}

	*plugins = dest

	for i, plugin := range *plugins {
		plugin.ID = fmt.Sprintf("%s.%d", plugin.Name, i)
		(*plugins)[i] = plugin
	}

	return nil
}

type Plugin struct {
	ID         string              `json:"-" yaml:"-"`
	Name       string              `json:"-" yaml:"-"`
	Properties property.Properties `json:",inline" yaml:",inline"`
}

func (p Plugin) MarshalJSON() ([]byte, error) {
	return p.Properties.MarshalJSON()
}

func (p Plugin) MarshalYAML() (any, error) {
	return p.Properties.MarshalYAML()
}

func (p *Plugin) UnmarshalJSON(data []byte) error {
	err := p.Properties.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	p.Name = Name(p.Properties)
	return nil
}

func (p *Plugin) UnmarshalYAML(node *yaml.Node) error {
	err := p.Properties.UnmarshalYAML(node)
	if err != nil {
		return err
	}

	p.Name = Name(p.Properties)
	return nil
}

func (p Plugin) Equal(target Plugin) bool {
	return p.Properties.Equal(target.Properties)
}

func (plugins Plugins) Equal(target Plugins) bool {
	return slices.EqualFunc(plugins, target, func(a, b Plugin) bool {
		return a.Equal(b)
	})
}
