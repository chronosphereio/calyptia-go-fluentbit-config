package fluentbitconfig

import (
	"encoding/json"
	"fmt"

	yaml "github.com/goccy/go-yaml"
	"golang.org/x/exp/slices"

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

func (plugins *Plugins) UnmarshalYAML(node []byte) error {
	var dest []Plugin
	err := yaml.Unmarshal(node, &dest)
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
	props, err := allPluginProperties(p, false /* classic */)
	if err != nil {
		return nil, fmt.Errorf("plugin %q: %w", p.ID, err)
	}

	return props.MarshalJSON()
}

func (p Plugin) MarshalYAML() (any, error) {
	props, err := allPluginProperties(p, false /* classic */)
	if err != nil {
		return nil, fmt.Errorf("plugin %q: %w", p.ID, err)
	}

	return props.MarshalYAML()
}

func (p *Plugin) UnmarshalJSON(data []byte) error {
	err := p.Properties.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	p.Name = Name(p.Properties)
	return nil
}

func (p *Plugin) UnmarshalYAML(node []byte) error {
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
