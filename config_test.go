package fluentbitconfig

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"gopkg.in/yaml.v3"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func Test_Config(t *testing.T) {
	names, err := filepath.Glob("testdata/*.conf")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var classicConf Config
			err = classicConf.UnmarshalClassic(classicText)
			assert.NoError(t, err)

			yamlText, err := os.ReadFile(strings.Replace(name, ".conf", ".yaml", 1))
			assert.NoError(t, err)

			var yamlConf Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			assert.NoError(t, err)

			jsonText, err := os.ReadFile(strings.Replace(name, ".conf", ".json", 1))
			assert.NoError(t, err)

			var jsonConf Config
			err = json.Unmarshal(jsonText, &jsonConf)
			assert.NoError(t, err)

			gotClassicText, err := yamlConf.DumpAsClassic()
			assert.NoError(t, err)
			assert.Equal(t, string(classicText), gotClassicText)

			gotYamlText, err := classicConf.DumpAsYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(yamlText), gotYamlText)

			var gotJsonText bytes.Buffer
			{
				enc := json.NewEncoder(&gotJsonText)
				enc.SetEscapeHTML(false)
				enc.SetIndent("", "    ")
				err = enc.Encode(classicConf)
				assert.NoError(t, err)
			}
			assert.NoError(t, err)
			assert.Equal(t, string(jsonText), gotJsonText.String())
		})
	}
}

func Test_Config_Processors(t *testing.T) {
	names, err := filepath.Glob("testdata/processors/*.yaml")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			yamlText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var yamlConf Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			assert.NoError(t, err)

			jsonText, err := os.ReadFile(strings.Replace(name, ".yaml", ".json", 1))
			assert.NoError(t, err)

			var jsonConf Config
			err = json.Unmarshal(jsonText, &jsonConf)
			assert.NoError(t, err)

			gotYamlText, err := yamlConf.DumpAsYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(yamlText), gotYamlText)

			var gotJsonText bytes.Buffer
			{
				enc := json.NewEncoder(&gotJsonText)
				enc.SetEscapeHTML(false)
				enc.SetIndent("", "    ")
				err = enc.Encode(yamlConf)
				assert.NoError(t, err)
			}
			assert.NoError(t, err)
			assert.Equal(t, string(jsonText), gotJsonText.String())
		})
	}
}

func makeTestName(fp string) string {
	s := filepath.Base(fp)
	s = strings.TrimRight(s, filepath.Ext(s))
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ReplaceAll(s, " ", "_")
}

func TestConfig_Equal(t *testing.T) {
	t.Run("equal_env", func(t *testing.T) {
		a := Config{
			Env: property.Properties{
				{Key: "foo", Value: "bar"},
			},
		}
		b := Config{
			Env: property.Properties{
				{Key: "foo", Value: "bar"},
			},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_env", func(t *testing.T) {
		a := Config{
			Env: property.Properties{
				{Key: "a", Value: "b"},
			},
		}
		b := Config{
			Env: property.Properties{
				{Key: "b", Value: "c"},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("equal_includes", func(t *testing.T) {
		a := Config{
			Includes: []string{"foo"},
		}
		b := Config{
			Includes: []string{"foo"},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_includes", func(t *testing.T) {
		a := Config{
			Includes: []string{"foo"},
		}
		b := Config{
			Includes: []string{"bar"},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("equal_input", func(t *testing.T) {
		a := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		b := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_input", func(t *testing.T) {
		a := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		b := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 22},
					}},
				},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("not_equal_properties_out_of_order", func(t *testing.T) {
		a := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "foo"},
						{Key: "test_key", Value: "test_value"},
					}},
				},
			},
		}
		b := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "test_key", Value: "test_value"},
						{Key: "name", Value: "foo"},
					}},
				},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("not_equal_by_names_out_of_order", func(t *testing.T) {
		a := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "one"},
					}}, {
					Properties: property.Properties{
						{Key: "name", Value: "two"},
					}},
				},
			},
		}
		b := Config{
			Pipeline: Pipeline{
				Inputs: Plugins{{
					Properties: property.Properties{
						{Key: "name", Value: "two"},
					}}, {
					Properties: property.Properties{
						{Key: "name", Value: "one"},
					}},
				},
			},
		}
		assert.False(t, a.Equal(b))
	})
}

func TestConfig_IDs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		conf, err := ParseAs(`
			[SERVICE]
				log_level error
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, nil, conf.IDs(true))
	})

	t.Run("ok", func(t *testing.T) {
		conf, err := ParseAs(`
			[INPUT]
				name tcp
			[OUTPUT]
				name  tcp
				match *
			[OUTPUT]
				name  stdout
				match *
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"input:tcp:tcp.0",
			"output:tcp:tcp.0",
			"output:stdout:stdout.1",
		}, conf.IDs(true))
	})

	t.Run("ok_2", func(t *testing.T) {
		conf, err := ParseAs(`
			[INPUT]
				name tcp
			[OUTPUT]
				name  stdout
				match *
			[OUTPUT]
				name  tcp
				match *
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"input:tcp:tcp.0",
			"output:stdout:stdout.0",
			"output:tcp:tcp.1",
		}, conf.IDs(true))
	})

	t.Run("ok_no_prefix", func(t *testing.T) {
		conf, err := ParseAs(`
 			[INPUT]
 				name tcp
 			[OUTPUT]
 				name  tcp
 				match *
 			[OUTPUT]
 				name  stdout
 				match *
 		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"tcp.0",
			"tcp.0",
			"stdout.1",
		}, conf.IDs(false))
	})
}

func TestConfig_FindByID(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		conf, err := ParseAs(`
			[INPUT]
				name cpu
			[INPUT]
				name mem
			[INPUT]
				name  cpu
		`, FormatClassic)
		assert.NoError(t, err)
		_, found := conf.FindByID("input:cpu:cpu.1")
		assert.False(t, found)
	})

	t.Run("ok", func(t *testing.T) {
		conf, err := ParseAs(`
			[INPUT]
				name cpu
			[INPUT]
				name mem
			[INPUT]
				name     cpu
				proptest valuetest
		`, FormatClassic)
		assert.NoError(t, err)
		plugin, found := conf.FindByID("input:cpu:cpu.2")
		assert.True(t, found)
		assert.Equal(t, Plugin{
			ID:   "cpu.2",
			Name: "cpu",
			Properties: []property.Property{
				{Key: "name", Value: "cpu"},
				{Key: "proptest", Value: "valuetest"},
			},
		}, plugin)
	})
}

func TestConfig_Validate_Classic(t *testing.T) {
	names, err := filepath.Glob("testdata/*.conf")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var classicConf Config
			err = classicConf.UnmarshalClassic(classicText)
			assert.NoError(t, err)
			classicConf.Validate()
		})
	}
}

func TestConfig_Validate_YAML(t *testing.T) {
	names, err := filepath.Glob("testdata/*.yaml")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			yamlText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var yamlConfig Config
			err = yaml.Unmarshal(yamlText, &yamlConfig)
			assert.NoError(t, err)
			yamlConfig.Validate()
		})
	}
}

func TestConfig_Validate_JSON(t *testing.T) {
	names, err := filepath.Glob("testdata/*.json")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			jsonText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var jsonConfig Config
			err = json.Unmarshal(jsonText, &jsonConfig)
			assert.NoError(t, err)
			jsonConfig.Validate()
		})
	}
}
