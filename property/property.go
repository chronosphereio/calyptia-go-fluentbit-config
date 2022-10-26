package property

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Properties list.
type Properties []Property

// MarshalYAML implements yaml.Marshaler interface.
func (pp Properties) MarshalYAML() (any, error) {
	if pp == nil {
		return nil, nil
	}

	if len(pp) == 0 {
		return []any{}, nil
	}

	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for _, p := range pp {
		valueNode := &yaml.Node{}
		if err := valueNode.Encode(p.Value); err != nil {
			return nil, fmt.Errorf("yaml encode property value: %w", err)
		}

		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: p.Key,
		}, valueNode)
	}

	return node, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
func (pp *Properties) UnmarshalYAML(node *yaml.Node) error {
	d := len(node.Content)
	if d%2 != 0 {
		return fmt.Errorf("expected even items for key-value")
	}

	for i := 0; i < d; i += 2 {
		var prop Property

		keyNode := node.Content[i]
		if err := keyNode.Decode(&prop.Key); err != nil {
			return fmt.Errorf("yaml decode property key: %w", err)
		}

		valueNode := node.Content[i+1]
		if err := valueNode.Decode(&prop.Value); err != nil {
			return fmt.Errorf("yaml decode property value: %w", err)
		}

		*pp = append(*pp, prop)
	}

	return nil
}

// Has property key insensitively.
func (pp Properties) Has(key string) bool {
	if pp == nil {
		return false
	}

	for _, p := range pp {
		if strings.EqualFold(p.Key, key) {
			return true
		}
	}

	return false
}

// Get property by its key insensitively.
func (pp Properties) Get(key string) (any, bool) {
	if pp == nil {
		return nil, false
	}

	for _, p := range pp {
		if strings.EqualFold(p.Key, key) {
			return p.Value, true
		}
	}

	return nil, false
}

// Set property. If it exists (case-insensitive) it will override it,
// otherwise it will just add it.
func (pp *Properties) Set(key string, value any) {
	if pp == nil {
		return
	}

	for i, p := range *pp {
		if strings.EqualFold(p.Key, key) {
			(*pp)[i].Value = value
			return
		}
	}

	pp.Add(key, value)
}

// Add property.
func (pp *Properties) Add(key string, value any) {
	if pp == nil {
		return
	}

	*pp = append(*pp, Property{Key: key, Value: value})
}

// Property key-value pair.
type Property struct {
	Key   string
	Value any
}
