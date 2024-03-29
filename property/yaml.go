package property

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// MarshalYAML implements yaml.Marshaler interface
// to marshall a sorted list of properties into an object.
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

// UnmarshalYAML implements yaml.Unmarshaler interface
// to unmarshal an object into a sorted list of properties.
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
