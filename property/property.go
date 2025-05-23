package property

import (
	"maps"
	"reflect"
	"slices"
	"strings"
)

// Properties list.
type Properties []Property

// Property key-value pair.
type Property struct {
	Key   string
	Value any
}

// AsMap output.
// It will merge properties with the same key into a slice.
func (pp *Properties) AsMap() map[string]any {
	if pp == nil {
		return nil
	}

	out := map[string]any{}
	for _, p := range *pp {
		if v, ok := out[p.Key]; ok {
			// If key already exists, we try to convert it to a slice
			// and append to it.
			if s, ok := v.([]any); ok {
				s = append(s, p.Value)
				out[p.Key] = s
			} else {
				out[p.Key] = []any{v, p.Value}
			}
		} else {
			out[p.Key] = p.Value
		}
	}
	return out
}

// Has property key insensitively.
func (pp *Properties) Has(key string) bool {
	if pp == nil {
		return false
	}

	for _, p := range *pp {
		if strings.EqualFold(p.Key, key) {
			return true
		}
	}

	return false
}

// Get property by its key insensitively.
func (pp *Properties) Get(key string) (any, bool) {
	if pp == nil {
		return nil, false
	}

	for _, p := range *pp {
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

// Add adds a property to the list.
func (pp *Properties) Add(key string, value any) {
	if pp == nil {
		return
	}

	*pp = append(*pp, Property{Key: key, Value: value})
}

// Equal returns true if pp and target have the same properties in the same order.
func (pp *Properties) Equal(target Properties) bool {
	props := Properties{}
	if pp != nil {
		props = *pp
	}

	return slices.EqualFunc(props, target, func(a, b Property) bool {
		return a.Key == b.Key && reflect.DeepEqual(a.Value, b.Value)
	})
}

// PropertiesFromMap returns the key-value pairs as Properties, ordered by key.
func PropertiesFromMap(kvs map[string]string) Properties {
	if len(kvs) == 0 {
		return nil
	}

	props := make(Properties, 0, len(kvs))
	for _, k := range slices.Sorted(maps.Keys(kvs)) {
		props = append(props, Property{Key: k, Value: kvs[k]})
	}

	return props
}
