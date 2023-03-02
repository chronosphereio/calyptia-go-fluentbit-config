package fluentbitconfig_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"gopkg.in/yaml.v3"

	fluentbitconfig "github.com/calyptia/go-fluentbit-config"
	"github.com/calyptia/go-fluentbit-config/property"
)

func Test_Config(t *testing.T) {
	names, err := filepath.Glob("testdata/*.conf")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var classicConf fluentbitconfig.Config
			err = classicConf.UnmarshalClassic(classicText)
			assert.NoError(t, err)

			yamlText, err := os.ReadFile(strings.Replace(name, ".conf", ".yaml", 1))
			assert.NoError(t, err)

			var yamlConf fluentbitconfig.Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			assert.NoError(t, err)

			jsonText, err := os.ReadFile(strings.Replace(name, ".conf", ".json", 1))
			assert.NoError(t, err)

			var jsonConf fluentbitconfig.Config
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

func makeTestName(fp string) string {
	s := filepath.Base(fp)
	s = strings.TrimRight(s, filepath.Ext(s))
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ReplaceAll(s, " ", "_")
}

func TestConfig_Equal(t *testing.T) {
	t.Run("equal_env", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Env: property.Properties{
				{Key: "foo", Value: "bar"},
			},
		}
		b := fluentbitconfig.Config{
			Env: property.Properties{
				{Key: "foo", Value: "bar"},
			},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_env", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Env: property.Properties{
				{Key: "a", Value: "b"},
			},
		}
		b := fluentbitconfig.Config{
			Env: property.Properties{
				{Key: "b", Value: "c"},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("equal_includes", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Includes: []string{"foo"},
		}
		b := fluentbitconfig.Config{
			Includes: []string{"foo"},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_includes", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Includes: []string{"foo"},
		}
		b := fluentbitconfig.Config{
			Includes: []string{"bar"},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("equal_input", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"dummy": property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		b := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"dummy": property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		assert.True(t, a.Equal(b))
	})

	t.Run("not_equal_input", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"dummy": property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 10},
					}},
				},
			},
		}
		b := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"dummy": property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "rate", Value: 22},
					}},
				},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("not_equal_properties_out_of_order", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"foo": property.Properties{
						{Key: "name", Value: "foo"},
						{Key: "test_key", Value: "test_value"},
					}},
				},
			},
		}
		b := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"foo": property.Properties{
						{Key: "test_key", Value: "test_value"},
						{Key: "name", Value: "foo"},
					}},
				},
			},
		}
		assert.False(t, a.Equal(b))
	})

	t.Run("not_equal_by_names_out_of_order", func(t *testing.T) {
		a := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"one": property.Properties{
						{Key: "name", Value: "one"},
					}}, {
					"two": property.Properties{
						{Key: "name", Value: "two"},
					}},
				},
			},
		}
		b := fluentbitconfig.Config{
			Pipeline: fluentbitconfig.Pipeline{
				Inputs: []fluentbitconfig.ByName{{
					"two": property.Properties{
						{Key: "name", Value: "two"},
					},
					"one": property.Properties{
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
		conf, err := fluentbitconfig.ParseAs(`
			[SERVICE]
				log_level error
		`, fluentbitconfig.FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, nil, conf.IDs(true))
	})

	t.Run("ok", func(t *testing.T) {
		conf, err := fluentbitconfig.ParseAs(`
			[INPUT]
				name tcp
			[OUTPUT]
				name  tcp
				match *
			[OUTPUT]
				name  stdout
				match *
		`, fluentbitconfig.FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"input:tcp:tcp.0",
			"output:tcp:tcp.0",
			"output:stdout:stdout.1",
		}, conf.IDs(true))
	})

	t.Run("ok_2", func(t *testing.T) {
		conf, err := fluentbitconfig.ParseAs(`
			[INPUT]
				name tcp
			[OUTPUT]
				name  stdout
				match *
			[OUTPUT]
				name  tcp
				match *
		`, fluentbitconfig.FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"input:tcp:tcp.0",
			"output:stdout:stdout.0",
			"output:tcp:tcp.1",
		}, conf.IDs(true))
	})

	t.Run("ok_no_prefix", func(t *testing.T) {
		conf, err := fluentbitconfig.ParseAs(`
			[INPUT]
				name tcp
			[OUTPUT]
				name  tcp
				match *
			[OUTPUT]
				name  stdout
				match *
		`, fluentbitconfig.FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"tcp.0",
			"tcp.0",
			"stdout.1",
		}, conf.IDs(false))
	})
}
