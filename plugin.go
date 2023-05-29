package fluentbitconfig

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

type Plugins []Plugin

// IDs not namespaced.
// For example: tail.0
func (plugins Plugins) IDs() []string {
	var ids []string
	for _, plugin := range plugins {
		ids = append(ids, plugin.ID)
	}
	return ids
}

// FindByID were the id should not be namespaced.
// For example: tail.0
func (plugins Plugins) FindByID(id string) (Plugin, bool) {
	for _, plugin := range plugins {
		if plugin.ID == id {
			return plugin, true
		}
	}
	return Plugin{}, false
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
	p.handleV1()

	return nil
}

func (p *Plugin) UnmarshalYAML(node *yaml.Node) error {
	err := p.Properties.UnmarshalYAML(node)
	if err != nil {
		return err
	}

	p.Name = Name(p.Properties)
	p.handleV1()

	return nil
}

// handleV1 adds support for old v1 format.
func (p *Plugin) handleV1() {
	if len(p.Properties) != 1 || p.Name != "" {
		return
	}

	for _, prop := range p.Properties {
		// TODO: keep order of properties.
		props, ok := prop.Value.(map[string]any)
		if !ok {
			break
		}

		p.Properties = property.Properties{}
		for k, v := range props {
			p.Properties.Add(k, v)
		}
		p.Name = Name(p.Properties)
	}
}

func (p Plugin) Equal(target Plugin) bool {
	return p.Properties.Equal(target.Properties)
}

func (plugins Plugins) Equal(target Plugins) bool {
	return slices.EqualFunc(plugins, target, func(a, b Plugin) bool {
		return a.Equal(b)
	})
}
