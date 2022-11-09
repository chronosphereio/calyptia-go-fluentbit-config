package fluentbitconfig

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validate with the default schema.
func (c Config) Validate() error {
	return c.ValidateWithSchema(DefaultSchema)
}

// ValidateWithSchema validates that the config satisfies
// all properties defined on the schema.
// That is if the schema says the the input plugin "cpu"
// has a property named "pid" that is of integer type,
// it must be a valid integer.
func (c Config) ValidateWithSchema(schema Schema) error {
	validate := func(kind string, byNames []ByName, schemaSections []SchemaSection) error {
		for _, byName := range byNames {
			for name, props := range byName {
				var foundPlugin bool
				for _, schemaSection := range schemaSections {
					if !strings.EqualFold(name, schemaSection.Name) {
						continue
					}

					for _, p := range props {
						if isCommonProperty(p.Key) {
							continue
						}

						if isCloudVariable(p.Value) {
							continue
						}

						var foundProp bool
						for _, schemaOptions := range schemaSection.Properties.All() {
							if !strings.EqualFold(p.Key, schemaOptions.Name) {
								continue
							}

							if !valid(schemaOptions, p.Value) {
								return fmt.Errorf("%s: %s: expected %q to be a valid %s, got %v", kind, name, p.Key, schemaOptions.Type, p.Value)
							}

							foundProp = true
							break
						}

						if !foundProp {
							return fmt.Errorf("%s: %s: unknown property %q", kind, name, p.Key)
						}
					}

					foundPlugin = true
					break
				}

				if !foundPlugin {
					return fmt.Errorf("%s: unknown plugin %q", kind, name)
				}
			}
		}

		// valid by default
		return nil
	}

	if err := validate("custom", c.Customs, schema.Customs); err != nil {
		return err
	}

	if err := validate("input", c.Pipeline.Inputs, schema.Inputs); err != nil {
		return err
	}

	if err := validate("filter", c.Pipeline.Filters, schema.Filters); err != nil {
		return err
	}

	if err := validate("output", c.Pipeline.Outputs, schema.Outputs); err != nil {
		return err
	}

	// valid by default
	return nil
}

func isCommonProperty(name string) bool {
	for _, got := range [...]string{"name", "alias", "tag", "match", "match_Regex"} {
		if strings.EqualFold(name, got) {
			return true
		}
	}
	return false
}

var (
	// reCloudSecretVariable matches `{{ secrets.example }}`
	reCloudSecretVariable = regexp.MustCompile(`{{\s*secrets\.\w+\s*}}`)
	// reCloudFileVariable matches `{{ files.example }}`
	reCloudFileVariable = regexp.MustCompile(`{{\s*files\.[0-9A-Za-z]+(?:-[A-Za-z]{3,4}|)+\s*}}`)
)

func isCloudVariable(val any) bool {
	s, ok := val.(string)
	if !ok {
		return false
	}

	return reCloudSecretVariable.MatchString(s) ||
		reCloudFileVariable.MatchString(s)
}

func valid(opts SchemaOptions, val any) bool {
	switch opts.Type {
	case "deprecated":
		// TODO: review whether "deprecated" is still valid and just ignored,
		// or totally invalid.
		return true
	case "string":
		_, ok := val.(string)
		return ok
	case "boolean":
		return validBoolean(val)
	case "integer":
		return validInteger(val)
	case "double":
		return validDouble(val)
	case "time":
		return validTime(val)
	case "size":
		return validByteSize(val)
	case "prefixed string":
		// TODO: stricter validator once we know the full rules.
		s, ok := val.(string)
		return ok && s != ""
	case "multiple comma delimited strings":
		_, ok := val.(string)
		return ok
	case "space delimited strings (minimum 1)":
		return validSpaceDelimitedString(val, 1)
	case "space delimited strings (minimum 2)":
		return validSpaceDelimitedString(val, 2)
	case "space delimited strings (minimum 3)":
		return validSpaceDelimitedString(val, 3)
	case "space delimited strings (minimum 4)":
		return validSpaceDelimitedString(val, 4)
	}

	// valid by default
	return true
}

func validBoolean(val any) bool {
	if s, ok := val.(string); ok {
		return strings.EqualFold(s, "true") ||
			strings.EqualFold(s, "false") ||
			strings.EqualFold(s, "on") ||
			strings.EqualFold(s, "off")
	}

	_, ok := val.(bool)
	return ok
}

func validInteger(val any) bool {
	if s, ok := val.(string); ok {
		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			return true
		}

		_, err := strconv.ParseUint(s, 10, 64)
		return err == nil
	}

	switch val.(type) {
	case int,
		int8, int16, int32, int64,
		uint8, uint16, uint32, uint64:
		return true
	}

	return false
}

func validDouble(val any) bool {
	if s, ok := val.(string); ok {
		_, err := strconv.ParseFloat(s, 64)
		return err == nil
	}

	switch val.(type) {
	case float32, float64:
		return true
	}

	return false
}

func validTime(val any) bool {
	if validInteger(val) || validDouble(val) {
		return true
	}

	// TODO: stricter time validator once we know the full rules.
	if s, ok := val.(string); ok {
		return s != ""
	}

	return false
}

// reByteSize example -45.32kB
var reByteSize = regexp.MustCompile(`^(?:\+|-)?[0-9]+(?:\.?[0-9]*)\s?(?:(?:k|K)|(?:m|M)|(?:g|G))?(?:b|B)?$`)

func validByteSize(val any) bool {
	// numeric bytes without unit.
	if validInteger(val) || validDouble(val) {
		return true
	}

	if s, ok := val.(string); ok {
		return reByteSize.MatchString(s)
	}

	return false
}

func validSpaceDelimitedString(val any, min int) bool {
	s, ok := val.(string)
	if !ok {
		return false
	}

	return len(strings.Fields(s)) >= min
}
