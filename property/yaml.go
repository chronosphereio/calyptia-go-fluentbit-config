package property

import (
	"fmt"

	"github.com/goccy/go-yaml"
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

	properties := make(yaml.MapSlice, len(pp))

	for pindex, prop := range pp {
		properties[pindex] = yaml.MapItem{Key: prop.Key, Value: prop.Value}
	}

	return properties, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
// to unmarshal an object into a sorted list of properties.
func (pp *Properties) UnmarshalYAML(node []byte) error {
	var mprops yaml.MapSlice

	if err := yaml.Unmarshal(node, &mprops); err != nil {
		return err
	}

	for _, mprop := range mprops {
		var prop Property
		var ok bool

		prop.Key, ok = mprop.Key.(string)
		if !ok {
			return fmt.Errorf("unable to unmarshal key")
		}
		prop.Value = mprop.Value

		*pp = append(*pp, prop)
	}

	return nil
}
