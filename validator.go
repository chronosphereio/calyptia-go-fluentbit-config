package fluentbitconfig

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/calyptia/go-fluentbit-config/v2/property"
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
	validate := func(kind SectionKind, plugins Plugins) error {
		for _, plugin := range plugins {
			if err := ValidateSectionWithSchema(kind, plugin.Properties, schema); err != nil {
				return err
			}
		}

		return nil
	}

	validateService := func(properties property.Properties) error {
		for _, property := range properties {
			switch property.Value.(type) {
			case bool:
			case int:
			case uint64:
			case int64:
			case float64:
			case string:
			default:
				return fmt.Errorf("invalid type (%T) for service setting: %s",
					property.Value,
					property.Key)
			}
		}
		return nil
	}

	if err := validate(SectionKindCustom, c.Customs); err != nil {
		return err
	}

	if err := validate(SectionKindInput, c.Pipeline.Inputs); err != nil {
		return err
	}

	if err := validate(SectionKindFilter, c.Pipeline.Filters); err != nil {
		return err
	}

	if err := validate(SectionKindOutput, c.Pipeline.Outputs); err != nil {
		return err
	}

	if err := validateService(c.Service); err != nil {
		return err
	}

	// valid by default
	return nil
}

func ValidateSection(kind SectionKind, props property.Properties) error {
	return ValidateSectionWithSchema(kind, props, DefaultSchema)
}

func ValidateSectionWithSchema(kind SectionKind, props property.Properties, schema Schema) error {
	name := Name(props)
	if name == "" {
		return ErrMissingName
	}

	// If the name takes a cloud variable, it won't be on the schema,
	// so we allow it.
	// TODO: review whether is valid that the `name` property can take a
	// cloud variable syntax.
	//
	// Example:
	// 	[INPUT]
	// 		Name {{files.myinput}}
	//
	// Maybe we can pass-over the actual value.
	if isCloudVariable(name) {
		return nil
	}

	section, ok := schema.findSection(kind, name)
	if !ok {
		return NewUnknownPluginError(kind, name)
	}

	for _, p := range props {
		if isCommonProperty(p.Key) || isCloudVariable(p.Key) || isCloudVariable(p.Value) || isCoreProperty(p.Key) {
			continue
		}

		if areProcessors(p.Key) {
			if err := validateProcessors(p.Value); err != nil {
				return err
			}
			continue
		}

		opts, ok := section.findOptions(p.Key)
		if !ok {
			return fmt.Errorf("%s: %s: unknown property %q", kind, name, p.Key)
		}

		if !valid(opts, p.Value) {
			return fmt.Errorf("%s: %s: expected %q to be a valid %s, got %v", kind, name, p.Key, opts.Type, p.Value)
		}
	}

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

func areProcessors(key string) bool {
	return strings.ToLower(key) == "processors"
}

// isCoreProperty tells whether the property is from Core.
// Core properties start with `core.` and are ignored.
// See https://github.com/chronosphereio/calyptia-core-fluent-bit/blob/71795c472bcb1f7e09a7ea121366d5f91df4263f/patches/pwhelan-flb_config-ignore-core-namespaced-properties.patch#L22-L25
func isCoreProperty(key string) bool {
	return strings.HasPrefix(key, "core.")
}

func validateProcessorsSectionMaps(sectionMaps []interface{}) error {
	for _, section := range sectionMaps {
		props := property.Properties{}
		for key, val := range section.(map[string]interface{}) {
			props.Set(key, val)
		}
		if err := ValidateSection(SectionKindFilter, props); err != nil {
			return err
		}

	}
	return nil
}

func validateProcessors(processors any) error {
	if procMap, ok := processors.(map[string]interface{}); !ok {
		return fmt.Errorf("not a list of processors")
	} else {
		for key, sections := range procMap {
			sectionMaps, ok := sections.([]interface{})
			if !ok {
				return fmt.Errorf("not a list of processors")
			}
			switch key {
			case "logs", "metrics", "traces":
				if err := validateProcessorsSectionMaps(sectionMaps); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown processor type: %s", key)
			}
		}
	}
	return nil
}

func valid(opts SchemaOptions, val any) bool {
	switch opts.Type {
	case "deprecated":
		// TODO: review whether "deprecated" is still valid and just ignored,
		// or totally invalid.
		return true
	case "string":
		// numbers and booleans are valid strings too.
		if validInteger(val) || validDouble(val) {
			return true
		}
		if _, ok := val.(bool); ok {
			return true
		}

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
		if _, ok := val.([]any); ok {
			for _, v := range val.([]any) {
				if _, ok := v.(string); !ok {
					return false
				}
			}

			return true
		}

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
	valid := func(val any, min int) bool {
		s, ok := val.(string)
		if !ok {
			return false
		}

		return len(strings.Fields(s)) >= min
	}

	if v, ok := val.([]any); ok {
		for _, s := range v {
			if !valid(s, min) {
				return false
			}
		}

		return true
	}

	return valid(val, min)
}
