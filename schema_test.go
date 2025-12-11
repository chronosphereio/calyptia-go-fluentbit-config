package fluentbitconfig

import (
	"fmt"
	"testing"
)

// Shared validation functions

// validateGetSchemaError validates that GetSchema returns an error for invalid inputs
func validateGetSchemaError(t *testing.T, version string, testName string) {
	t.Helper()
	schema, err := GetSchema(version)
	if err == nil {
		t.Errorf("%s: GetSchema(%s) expected error but got none", testName, version)
	}
	// Ensure schema is empty when there's an error
	if schema.FluentBit.Version != "" {
		t.Errorf("%s: Expected empty schema on error, but got version %s", testName, schema.FluentBit.Version)
	}
}

// validateSchemaOptionsMatch validates that a found option matches expected values
func validateSchemaOptionsMatch(t *testing.T, found SchemaOptions, expected SchemaOptions, testName string) {
	t.Helper()
	if found.Name != expected.Name {
		t.Errorf("%s: Expected option name %s, got %s", testName, expected.Name, found.Name)
	}
	if found.Type != expected.Type {
		t.Errorf("%s: Expected option type %s, got %s", testName, expected.Type, found.Type)
	}
	if found.Description != expected.Description {
		t.Errorf("%s: Expected option description %s, got %s", testName, expected.Description, found.Description)
	}
	if found.Default != expected.Default {
		t.Errorf("%s: Expected option default %v, got %v", testName, expected.Default, found.Default)
	}
}

// validateSchemaSectionMatch validates that a found section matches expected values
func validateSchemaSectionMatch(t *testing.T, found SchemaSection, expected SchemaSection, testName string) {
	t.Helper()
	if found.Type != expected.Type {
		t.Errorf("%s: Expected section type %s, got %s", testName, expected.Type, found.Type)
	}
	if found.Name != expected.Name {
		t.Errorf("%s: Expected section name %s, got %s", testName, expected.Name, found.Name)
	}
	if found.Description != expected.Description {
		t.Errorf("%s: Expected section description %s, got %s", testName, expected.Description, found.Description)
	}
}

// validateFindSectionsResult validates the result of findSections
func validateFindSectionsResult(t *testing.T, sections []SchemaSection, found bool, expectedFound bool, expectedCount int, testName string) {
	t.Helper()
	if found != expectedFound {
		t.Errorf("%s: Expected found=%v, got found=%v", testName, expectedFound, found)
	}
	if len(sections) != expectedCount {
		t.Errorf("%s: Expected %d sections, got %d", testName, expectedCount, len(sections))
	}
}

// validateLTSPluginDetails validates specific LTS plugin configuration
func validateLTSPluginDetails(t *testing.T, plugin SchemaSection, expectedType, expectedDescription string, expectedOptionCount int, testName string) {
	t.Helper()
	if plugin.Type != expectedType {
		t.Errorf("%s: Expected type %s, got %s", testName, expectedType, plugin.Type)
	}
	if plugin.Description != expectedDescription {
		t.Errorf("%s: Expected description %s, got %s", testName, expectedDescription, plugin.Description)
	}
	if len(plugin.Properties.Options) != expectedOptionCount {
		t.Errorf("%s: Expected %d options, got %d", testName, expectedOptionCount, len(plugin.Properties.Options))
	}
}

// validatePluginOption validates a specific option within a plugin
func validatePluginOption(t *testing.T, plugin SchemaSection, optionName, expectedType string, testName string) {
	t.Helper()
	option, found := plugin.findOptions(optionName)
	if !found {
		t.Errorf("%s: Expected option %s not found", testName, optionName)
		return
	}
	if option.Type != expectedType {
		t.Errorf("%s: Expected option type %s, got %s", testName, expectedType, option.Type)
	}
}

// validatePluginDefaultValue validates a specific option's default value
func validatePluginDefaultValue(t *testing.T, plugin SchemaSection, optionName string, expectedDefault interface{}, testName string) {
	t.Helper()
	option, found := plugin.findOptions(optionName)
	if !found {
		t.Fatalf("%s: Option %s not found in plugin", testName, optionName)
	}
	if option.Default != expectedDefault {
		t.Errorf("%s: Expected default value %v, got %v", testName, expectedDefault, option.Default)
	}
}

// findInputPluginInSchema finds a plugin by name in the schema inputs
func findInputPluginInSchema(schema Schema, pluginName string) (SchemaSection, bool) {
	for _, input := range schema.Inputs {
		if input.Name == pluginName {
			return input, true
		}
	}
	return SchemaSection{}, false
}

// findProcessorPluginInSchema finds a plugin by name in the schema inputs
func findProcessorPluginInSchema(schema Schema, pluginName string) (SchemaSection, bool) {
	for _, input := range schema.Processors {
		if input.Name == pluginName {
			return input, true
		}
	}
	return SchemaSection{}, false
}

// Test functions

func TestDefaultSchema(t *testing.T) {
	// Test that the DefaultSchema is properly initialized
	if DefaultSchema.FluentBit.Version == "" {
		t.Error("DefaultSchema should have a version")
	}

	// Test that LTS plugins are injected
	if len(DefaultSchema.Inputs) == 0 {
		t.Error("DefaultSchema should have input plugins")
	}

	// Verify that at least one of the LTS plugins is present
	found := false
	for _, input := range DefaultSchema.Inputs {
		if input.Name == "s3_sqs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DefaultSchema should contain LTS plugins after InjectLTSPlugins()")
	}
}

// GetSchema tests

func TestGetSchema_InvalidSemanticVersion(t *testing.T) {
	validateGetSchemaError(t, "invalid", "invalid semantic version")
}

func TestGetSchema_NonExistentVersion(t *testing.T) {
	validateGetSchemaError(t, "999.999.999", "non-existent version")
}

func TestGetSchema_EmptyVersion(t *testing.T) {
	validateGetSchemaError(t, "", "empty version")
}

func TestGetSchema_VersionWithVPrefix(t *testing.T) {
	validateGetSchemaError(t, "v1.0.0", "version with v prefix")
}

func TestSchemaProperties_All(t *testing.T) {
	props := SchemaProperties{
		Options: []SchemaOptions{
			{Name: "option1", Type: "string"},
			{Name: "option2", Type: "integer"},
		},
		GlobalOptions: []SchemaOptions{
			{Name: "global1", Type: "boolean"},
		},
		Networking: []SchemaOptions{
			{Name: "network1", Type: "string"},
		},
		NetworkTLS: []SchemaOptions{
			{Name: "tls1", Type: "string"},
		},
	}

	all := props.all()

	expectedCount := len(props.Options) + len(props.GlobalOptions) + len(props.Networking) + len(props.NetworkTLS)
	if len(all) != expectedCount {
		t.Errorf("Expected %d options, got %d", expectedCount, len(all))
	}

	// Verify all options are present
	names := make(map[string]bool)
	for _, opt := range all {
		names[opt.Name] = true
	}

	expectedNames := []string{"option1", "option2", "global1", "network1", "tls1"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Expected option %s not found in all()", name)
		}
	}
}

// SchemaSection.findOptions tests

func TestSchemaSection_FindOptions_ExistingOption(t *testing.T) {
	section := SchemaSection{
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{Name: "test_option", Type: "string", Description: "Test option"},
			},
		},
	}

	option, found := section.findOptions("test_option")
	if !found {
		t.Error("Expected to find existing option")
	}
	expected := SchemaOptions{Name: "test_option", Type: "string", Description: "Test option"}
	validateSchemaOptionsMatch(t, option, expected, "find existing option")
}

func TestSchemaSection_FindOptions_CaseInsensitive(t *testing.T) {
	section := SchemaSection{
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{Name: "UPPER_OPTION", Type: "integer", Description: "Upper case option"},
			},
		},
	}

	option, found := section.findOptions("UPPER_option")
	if !found {
		t.Error("Expected to find case insensitive option")
	}
	expected := SchemaOptions{Name: "UPPER_OPTION", Type: "integer", Description: "Upper case option"}
	validateSchemaOptionsMatch(t, option, expected, "find case insensitive")
}

func TestSchemaSection_FindOptions_InGlobalOptions(t *testing.T) {
	section := SchemaSection{
		Properties: SchemaProperties{
			GlobalOptions: []SchemaOptions{
				{Name: "global_option", Type: "boolean", Description: "Global option"},
			},
		},
	}

	option, found := section.findOptions("global_option")
	if !found {
		t.Error("Expected to find option in global options")
	}
	expected := SchemaOptions{Name: "global_option", Type: "boolean", Description: "Global option"}
	validateSchemaOptionsMatch(t, option, expected, "find in global options")
}

func TestSchemaSection_FindOptions_NotFound(t *testing.T) {
	section := SchemaSection{
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{Name: "existing_option", Type: "string"},
			},
		},
	}

	_, found := section.findOptions("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent option")
	}
}

// Schema.findSection tests

func TestSchema_FindSection_InputSection(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "dummy", Description: "Dummy input"},
		},
	}

	section, found := schema.findSection(SectionKindInput, "dummy")
	if !found {
		t.Error("Expected to find input section")
	}
	expected := SchemaSection{Type: "input", Name: "dummy", Description: "Dummy input"}
	validateSchemaSectionMatch(t, section, expected, "find input section")
}

func TestSchema_FindSection_CaseInsensitive(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "TAIL", Description: "Tail input"},
		},
	}

	section, found := schema.findSection(SectionKindInput, "tail")
	if !found {
		t.Error("Expected to find case insensitive section")
	}
	expected := SchemaSection{Type: "input", Name: "TAIL", Description: "Tail input"}
	validateSchemaSectionMatch(t, section, expected, "find case insensitive")
}

func TestSchema_FindSection_OutputSection(t *testing.T) {
	schema := Schema{
		Outputs: []SchemaSection{
			{Type: "output", Name: "stdout", Description: "Standard output"},
		},
	}

	section, found := schema.findSection(SectionKindOutput, "stdout")
	if !found {
		t.Error("Expected to find output section")
	}
	expected := SchemaSection{Type: "output", Name: "stdout", Description: "Standard output"}
	validateSchemaSectionMatch(t, section, expected, "find output section")
}

func TestSchema_FindSection_FilterSection(t *testing.T) {
	schema := Schema{
		Filters: []SchemaSection{
			{Type: "filter", Name: "grep", Description: "Grep filter"},
		},
	}

	section, found := schema.findSection(SectionKindFilter, "grep")
	if !found {
		t.Error("Expected to find filter section")
	}
	expected := SchemaSection{Type: "filter", Name: "grep", Description: "Grep filter"}
	validateSchemaSectionMatch(t, section, expected, "find filter section")
}

func TestSchema_FindSection_CustomSection(t *testing.T) {
	schema := Schema{
		Customs: []SchemaSection{
			{Type: "custom", Name: "custom1", Description: "Custom section"},
		},
	}

	section, found := schema.findSection(SectionKindCustom, "custom1")
	if !found {
		t.Error("Expected to find custom section")
	}
	expected := SchemaSection{Type: "custom", Name: "custom1", Description: "Custom section"}
	validateSchemaSectionMatch(t, section, expected, "find custom section")
}

func TestSchema_FindSection_NotFound(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "dummy"},
		},
	}

	_, found := schema.findSection(SectionKindInput, "nonexistent")
	if found {
		t.Error("Expected not to find nonexistent section")
	}
}

func TestSchema_FindSection_InvalidKind(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "dummy"},
		},
	}

	_, found := schema.findSection(SectionKind("invalid"), "dummy")
	if found {
		t.Error("Expected not to find section with invalid kind")
	}
}

// Schema.findSections tests

func TestSchema_FindSections_InputSections(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "dummy"},
			{Type: "input", Name: "tail"},
		},
	}

	sections, found := schema.findSections(SectionKindInput)
	validateFindSectionsResult(t, sections, found, true, 2, "find input sections")
}

func TestSchema_FindSections_OutputSections(t *testing.T) {
	schema := Schema{
		Outputs: []SchemaSection{
			{Type: "output", Name: "stdout"},
		},
	}

	sections, found := schema.findSections(SectionKindOutput)
	validateFindSectionsResult(t, sections, found, true, 1, "find output sections")
}

func TestSchema_FindSections_FilterSections(t *testing.T) {
	schema := Schema{
		Filters: []SchemaSection{
			{Type: "filter", Name: "grep"},
		},
	}

	sections, found := schema.findSections(SectionKindFilter)
	validateFindSectionsResult(t, sections, found, true, 1, "find filter sections")
}

func TestSchema_FindSections_CustomSections(t *testing.T) {
	schema := Schema{
		Customs: []SchemaSection{
			{Type: "custom", Name: "custom1"},
		},
	}

	sections, found := schema.findSections(SectionKindCustom)
	validateFindSectionsResult(t, sections, found, true, 1, "find custom sections")
}

func TestSchema_FindSections_InvalidKind(t *testing.T) {
	schema := Schema{}

	sections, found := schema.findSections(SectionKind("invalid"))
	validateFindSectionsResult(t, sections, found, false, 0, "invalid section kind")
}

func TestSchema_FindSections_ParserKindNotHandled(t *testing.T) {
	schema := Schema{}

	sections, found := schema.findSections(SectionKindParser)
	validateFindSectionsResult(t, sections, found, false, 0, "parser section kind not handled")
}

func TestSchema_InjectLTSPlugins(t *testing.T) {
	schema := Schema{
		Inputs: []SchemaSection{
			{Type: "input", Name: "existing", Description: "Existing input"},
		},
	}

	originalCount := len(schema.Inputs)
	schema.InjectLTSPlugins()

	// Should have more inputs after injection
	if len(schema.Inputs) <= originalCount {
		t.Error("InjectLTSPlugins should add LTS plugins to inputs")
	}

	// Test that specific LTS plugins are present
	expectedLTSPlugins := []string{
		"aws_kinesis_stream",
		"azeventgrid",
		"azure-blob-input",
		"cloudflare",
		"datagen",
		"gcp_pubsub",
		"gdummy",
		"go-s3-replay-plugin",
		"gsuite-reporter",
		"http_loader",
		"http_scraper",
		"s3_sqs",
		"sqldb",
	}

	pluginMap := make(map[string]bool)
	for _, input := range schema.Inputs {
		pluginMap[input.Name] = true
	}

	for _, expectedPlugin := range expectedLTSPlugins {
		if !pluginMap[expectedPlugin] {
			t.Errorf("Expected LTS plugin %s not found after injection", expectedPlugin)
		}
	}

	// Test that original plugin is still present
	if !pluginMap["existing"] {
		t.Error("Original input plugin should still be present after LTS injection")
	}
}

// LTS Plugin specific details tests

func TestSchema_InjectLTSPlugins_S3SQS(t *testing.T) {
	schema := Schema{}
	schema.InjectLTSPlugins()

	plugin, found := findInputPluginInSchema(schema, "s3_sqs")
	if !found {
		t.Fatal("Plugin s3_sqs not found")
	}

	validateLTSPluginDetails(t, plugin, "input", "Calyptia LTS advanced plugin providing logs replay from sqs events", 22, "s3_sqs details")
	validatePluginOption(t, plugin, "aws_bucket_name", "string", "s3_sqs aws_bucket_name option check")
	validatePluginOption(t, plugin, "sqs_queue_name", "string", "s3_sqs sqs_queue_name option check")
	validatePluginDefaultValue(t, plugin, "delete_messages", "true", "s3_sqs delete_messages default")
	validatePluginDefaultValue(t, plugin, "match_regexp", ".*", "s3_sqs match_regexp default")
}

func TestSchema_InjectLTSPlugins_Calyptia(t *testing.T) {
	schema := Schema{}
	schema.InjectLTSPlugins()

	plugin, found := findProcessorPluginInSchema(schema, "calyptia")
	if !found {
		t.Fatal("Plugin calyptia not found")
	}

	validateLTSPluginDetails(t, plugin,
		"processor", "calyptia actions processor", 1, "calyptia processor details")
	validatePluginOption(t, plugin, "actions", "multiple keyvalues",
		"calyptia processor actions check")

	actions, _ := plugin.findOptions("actions")
	fmt.Printf("actions=%+v\n", actions)
}

func TestSchema_InjectLTSPlugins_GCPPubSub(t *testing.T) {
	schema := Schema{}
	schema.InjectLTSPlugins()

	plugin, found := findInputPluginInSchema(schema, "gcp_pubsub")
	if !found {
		t.Fatal("Plugin gcp_pubsub not found")
	}

	validateLTSPluginDetails(t, plugin, "input",
		"Google Cloud Pub/Sub input plugin for Cloud Logging LogEntry messages.", 3,
		"gcp_pubsub details")

	validatePluginOption(t, plugin, "project_id", "string", "gcp_pubsub project_id option check")
	validatePluginOption(t, plugin, "subscription_id", "string", "gcp_pubsub subscription_id option check")
	validatePluginOption(t, plugin, "endpoint", "string", "gcp_pubsub endpoint option check")
}
