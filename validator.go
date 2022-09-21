package fluentbit_config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
)

//go:embed schemas/1-9-7-with-ports.json
var rawSchema []byte

var DefaultSchema = func() Schema {
	var s Schema
	err := json.Unmarshal(rawSchema, &s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse fluent-bit schema: %v\n", err)
		os.Exit(1)
	}

	s.InjectLTSPlugins()

	return s
}()

func (s *Schema) InjectLTSPlugins() {
	// See https://github.com/calyptia/core-images/blob/main/container/plugins/plugins.json
	s.Inputs = append(s.Inputs, SchemaSection{
		// See https://github.com/calyptia/lts-advanced-plugin-s3-replay
		Type:        "input",
		Name:        "go-s3-replay-plugin",
		Description: "Calyptia LTS advanced plugin providing logs replay from s3",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name: "aws_access_key",
					Type: "string",
				},
				{
					Name: "aws_secret_key",
					Type: "string",
				},
				{
					Name: "aws_bucket_name",
					Type: "string",
				},
				{
					Name:        "aws_s3_endpoint",
					Type:        "string",
					Description: "Either aws_s3_endpoint or aws_bucket_region has to be provided",
				},
				{
					Name:        "aws_bucket_region",
					Type:        "string",
					Description: "Either aws_s3_endpoint or aws_bucket_region has to be provided",
				},
				{
					Name:        "logs",
					Type:        "string",
					Description: "Log pattern",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/calyptia/lts-advanced-plugin-dummy
		Type:        "input",
		Name:        "gdummy",
		Description: "dummy GO!",
	}, SchemaSection{
		// See https://github.com/calyptia/lts-advanced-plugin-gsuite-reporter
		Type:        "input",
		Name:        "gsuite-reporter",
		Description: "A Calyptia LTS advanced plugin providing activity streams from Gsuite",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "access_token",
					Type:        "string",
					Description: "Set either access_token or creds_file",
				},
				{
					Name:        "creds_file",
					Type:        "string",
					Description: "Service account or refresh token. It must set subject if this is not empty",
				},
				{
					Name:        "subject",
					Type:        "string",
					Description: "Email of the user to impersonate",
				},
				{
					Name:        "pull_interval",
					Type:        "string",
					Description: "Must be >= 1s",
					Default:     "30s",
				},
				{
					Name:        "data_dir",
					Type:        "string",
					Description: "Directory to store data. It holds a 1MB cache",
					Default:     "/data/storage",
				},
				{
					Name:        "telemetry",
					Type:        "boolean",
					Description: "Enable telemetry on Gsuite API",
				},
				{
					Name:        "application_name",
					Type:        "string",
					Description: "See https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities/list#ApplicationName",
				},
				{
					Name:        "user_key",
					Type:        "string",
					Description: "Profile ID or the user email for which the data should be filtered. Can be `all` for all information, or `userKey` for a user's unique Google Workspace profile ID or their primary email address",
					Default:     "all",
				},
			},
		},
	})
}

type Schema struct {
	FluentBit SchemaFluentBit `json:"fluent-bit"`
	Customs   []SchemaSection `json:"customs"`
	Inputs    []SchemaSection `json:"inputs"`
	Filters   []SchemaSection `json:"filters"`
	Outputs   []SchemaSection `json:"outputs"`
}

type SchemaFluentBit struct {
	Version       string `json:"version"`
	SchemaVersion string `json:"schema_version"`
	OS            string `json:"os"`
}

type SchemaSection struct {
	Type        string           `json:"type"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Properties  SchemaProperties `json:"properties"`
}

type SchemaProperties struct {
	Options []SchemaOptions `json:"options"`

	// Networking is only present in outputs.
	Networking []SchemaOptions `json:"networking"`
	// NetworkTLS is only present in outputs.
	NetworkTLS []SchemaOptions `json:"network_tls"`
}

type SchemaOptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
	Type        string `json:"type"`
}

// Validate with the default schema.
func (cfg *Config) Validate() error {
	return cfg.ValidateWithSchema(DefaultSchema)
}

// ValidateWithSchema validates that the config satisfies
// all properties defined on the schema.
// That is if the schema says the the input plugin "cpu"
// has a property named "pid" that is of integer type,
// it must be a valid integer.
func (cfg *Config) ValidateWithSchema(schema Schema) error {
	for _, section := range cfg.Sections {
		if !supportedSchemaSection(section.Type) {
			// TODO: support all sections on schema.
			continue
		}

		if err := section.ValidateWithSchema(schema); err != nil {
			return err
		}
	}

	return nil
}

func supportedSchemaSection(typ ConfigSectionType) bool {
	list := [...]ConfigSectionType{CustomSection, InputSection, FilterSection, OutputSection}
	var supported bool
	for _, got := range list {
		if typ == got {
			supported = true
			break
		}
	}

	return supported
}

func findSchemaSections(schema Schema, typ ConfigSectionType) ([]SchemaSection, bool) {
	switch typ {
	case CustomSection:
		return schema.Customs, true
	case InputSection:
		return schema.Inputs, true
	case FilterSection:
		return schema.Filters, true
	case OutputSection:
		return schema.Outputs, true
	}

	return nil, false
}

func findSchemaSection(ss []SchemaSection, pluginName string) (SchemaSection, bool) {
	for _, section := range ss {
		if strings.EqualFold(section.Name, pluginName) {
			return section, true
		}
	}

	return SchemaSection{}, false
}

func (section ConfigSection) Validate() error {
	return section.ValidateWithSchema(DefaultSchema)
}

func (section ConfigSection) ValidateWithSchema(schema Schema) error {
	if !supportedSchemaSection(section.Type) {
		// TODO: support all sections on schema.
		return fmt.Errorf("unsupported plugin kind %q", ConfigSectionTypes[section.Type])
	}

	schemaSections, ok := findSchemaSections(schema, section.Type)
	if !ok {
		return fmt.Errorf("unknown section %q", ConfigSectionTypes[section.Type])
	}

	pluginName := getPluginNameParameter(section)
	schemaSection, ok := findSchemaSection(schemaSections, pluginName)
	if !ok {
		return fmt.Errorf("unknown plugin %q", pluginName)
	}

	for _, field := range section.Fields {
		if isCloudVariable(field.Values) {
			continue
		}

		if isCommonProperty(field.Key) {
			if len(field.Values) == 0 {
				return fmt.Errorf("%s: expected %q to be a valid string; got %q", schemaSection.Name, field.Key, field.Values.ToString())
			}

			continue
		}

		opts, ok := findSchemaOptions(schemaSection.Properties, field.Key)
		if !ok {
			return fmt.Errorf("%s: unknown property %q", schemaSection.Name, field.Key)
		}

		if !validProp(opts, field) {
			return fmt.Errorf("%s: expected %q to be a valid %s; got %q", schemaSection.Name, opts.Name, opts.Type, field.Values.ToString())
		}
	}

	return nil
}

func isCommonProperty(name string) bool {
	for _, got := range [...]string{"Name", "Alias", "Tag", "Match", "Match_Regex"} {
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

func isCloudVariable(vals Values) bool {
	s := vals.ToString()
	return reCloudSecretVariable.MatchString(s) ||
		reCloudFileVariable.MatchString(s)
}

func findSchemaOptions(pp SchemaProperties, propertyName string) (SchemaOptions, bool) {
	var out SchemaOptions

	for _, opts := range pp.Options {
		if strings.EqualFold(opts.Name, propertyName) {
			return opts, true
		}
	}

	for _, opts := range pp.Networking {
		if strings.EqualFold(opts.Name, propertyName) {
			return opts, true
		}
	}

	for _, opts := range pp.NetworkTLS {
		if strings.EqualFold(opts.Name, propertyName) {
			return opts, true
		}
	}

	return out, false
}

func validProp(opts SchemaOptions, prop Field) bool {
	switch opts.Type {
	case "deprecated":
		// TODO: review whether "deprecated" is still valid and just ignored,
		// or totally invalid.
		return true
	case "string":
		return len(prop.Values) != 0
	case "boolean":
		return validBoolean(prop)
	case "integer":
		return validInteger(prop)
	case "double":
		return validDouble(prop)
	case "time":
		// TODO: stricter time validator once we know the full rules.
		return len(prop.Values) != 0
	case "size":
		return validByteSize(prop)
	case "prefixed string":
		// TODO: stricter time validator once we know the full rules.
		return len(prop.Values) != 0
	case "multiple comma delimited strings":
		return len(prop.Values) != 0
	case "space delimited strings (minimum 1)":
		return validSpaceDelimitedString(prop, 1)
	case "space delimited strings (minimum 2)":
		return validSpaceDelimitedString(prop, 2)
	case "space delimited strings (minimum 3)":
		return validSpaceDelimitedString(prop, 3)
	case "space delimited strings (minimum 4)":
		return validSpaceDelimitedString(prop, 4)
	}

	// valid by default
	return true
}

func validBoolean(prop Field) bool {
	if len(prop.Values) != 1 {
		return false
	}
	val := prop.Values[0]
	// it could have been caught as a string.
	// mostly due to "on" and "off" being treated as booleans too.
	if val.String != nil {
		s := *val.String
		return strings.EqualFold(s, "true") ||
			strings.EqualFold(s, "false") ||
			strings.EqualFold(s, "on") ||
			strings.EqualFold(s, "off")
	}
	return val.Bool != nil
}

func validInteger(prop Field) bool {
	if len(prop.Values) != 1 {
		return false
	}
	val := prop.Values[0]
	// it could have been caught as a string.
	if val.String != nil {
		_, err := strconv.ParseInt(*val.String, 10, 64)
		return err == nil
	}
	return val.Number != nil
}

func validDouble(prop Field) bool {
	if len(prop.Values) != 1 {
		return false
	}
	val := prop.Values[0]
	// it could have been caught as a string.
	if val.String != nil {
		_, err := strconv.ParseFloat(*val.String, 64)
		return err == nil
	}
	return val.Float != nil
}

func validByteSize(prop Field) bool {
	if len(prop.Values) == 0 {
		return false
	}
	// plain numeric bytes without unit.
	if len(prop.Values) == 1 && prop.Values[0].Number != nil {
		return true
	}

	// try with number as string
	if len(prop.Values) == 1 && prop.Values[0].String != nil {
		_, err := strconv.ParseInt(*prop.Values[0].String, 10, 64)
		return err == nil
	}

	// try with bytes parser.
	_, err := bytefmt.ToBytes(prop.Values.ToString())
	if err == nil {
		return true
	}

	// case when numer and unit were split by space.
	if len(prop.Values) == 2 && prop.Values[0].Number != nil && prop.Values[1].Unit != nil {
		return true
	}

	// case when numer and unit were split by space, but unit was detected as string.
	if len(prop.Values) == 2 && prop.Values[0].Number != nil && prop.Values[1].String != nil {
		return true
	}

	return len(prop.Values) == 1 && prop.Values[0].Unit != nil
}

func validSpaceDelimitedString(prop Field, min int) bool {
	return len(strings.Fields(prop.Values.ToString())) >= min
}
