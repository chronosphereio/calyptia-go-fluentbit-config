package fluentbitconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func Test_Config(t *testing.T) {
	names, err := filepath.Glob("testdata/*.conf")
	require.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			require.NoError(t, err)

			var classicConf Config
			err = classicConf.UnmarshalClassic(classicText)
			require.NoError(t, err)

			yamlText, err := os.ReadFile(strings.Replace(name, ".conf", ".yaml", 1))
			require.NoError(t, err)

			var yamlConf Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			require.NoError(t, err)

			jsonText, err := os.ReadFile(strings.Replace(name, ".conf", ".json", 1))
			require.NoError(t, err)

			var jsonConf Config
			err = json.Unmarshal(jsonText, &jsonConf)
			require.NoError(t, err)

			gotClassicText, err := yamlConf.DumpAsClassic()
			require.NoError(t, err)
			require.Equal(t, string(classicText), gotClassicText)

			gotYamlText, err := classicConf.DumpAsYAML()
			require.NoError(t, err)
			require.Equal(t, string(yamlText), gotYamlText)

			var gotJsonText bytes.Buffer
			{
				enc := json.NewEncoder(&gotJsonText)
				enc.SetEscapeHTML(false)
				enc.SetIndent("", "    ")
				err = enc.Encode(classicConf)
				require.NoError(t, err)
			}
			require.NoError(t, err)
			require.Equal(t, string(jsonText), gotJsonText.String())
		})
	}
}

func Test_Config_Processors(t *testing.T) {
	names, err := filepath.Glob("testdata/processors/*.yaml")
	require.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			yamlText, err := os.ReadFile(name)
			require.NoError(t, err)

			var yamlConf Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			require.NoError(t, err)

			jsonText, err := os.ReadFile(strings.Replace(name, ".yaml", ".json", 1))
			require.NoError(t, err)

			var jsonConf Config
			err = json.Unmarshal(jsonText, &jsonConf)
			require.NoError(t, err)

			gotYamlText, err := yamlConf.DumpAsYAML()
			require.NoError(t, err)
			require.Equal(t, string(yamlText), gotYamlText)

			var gotJsonText bytes.Buffer
			{
				enc := json.NewEncoder(&gotJsonText)
				enc.SetEscapeHTML(false)
				enc.SetIndent("", "    ")
				err = enc.Encode(yamlConf)
				require.NoError(t, err)
			}
			require.NoError(t, err)
			require.Equal(t, string(jsonText), gotJsonText.String())
		})
	}
}

func Test_Config_Parsers(t *testing.T) {
	cfg := Config{
		Pipeline: Pipeline{
			Inputs: Plugins{{
				Properties: property.Properties{
					{Key: "name", Value: "dummy"},
				},
			}},
			Outputs: Plugins{{
				Properties: property.Properties{
					{Key: "name", Value: "stdout"},
				},
			}},
		},
		Parsers: Plugins{{
			Properties: property.Properties{
				{Key: "name", Value: "myparser"},
				{Key: "key_name", Value: "log"},
				{Key: "format", Value: "regex"},
				{Key: "regex", Value: "^(?<foo>foo) bar$"},
			},
		}},
	}

	gotYamlText, err := cfg.DumpAsYAML()
	require.NoError(t, err)
	require.Equal(t, `pipeline:
    inputs:
        - name: dummy
    outputs:
        - name: stdout
parsers:
    - name: myparser
      key_name: log
      format: regex
      regex: ^(?<foo>foo) bar$
`, gotYamlText)
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
		require.True(t, a.Equal(b))
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
		require.False(t, a.Equal(b))
	})

	t.Run("equal_includes", func(t *testing.T) {
		a := Config{
			Includes: []string{"foo"},
		}
		b := Config{
			Includes: []string{"foo"},
		}
		require.True(t, a.Equal(b))
	})

	t.Run("not_equal_includes", func(t *testing.T) {
		a := Config{
			Includes: []string{"foo"},
		}
		b := Config{
			Includes: []string{"bar"},
		}
		require.False(t, a.Equal(b))
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
		require.True(t, a.Equal(b))
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
		require.False(t, a.Equal(b))
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
		require.False(t, a.Equal(b))
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
		require.False(t, a.Equal(b))
	})
}

func TestConfig_IDs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		conf, err := ParseAs(`
			[SERVICE]
				log_level error
		`, FormatClassic)
		require.NoError(t, err)
		require.Nil(t, conf.IDs(true))
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
		require.NoError(t, err)
		require.Equal(t, []string{
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
		require.NoError(t, err)
		require.Equal(t, []string{
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
		require.NoError(t, err)
		require.Equal(t, []string{
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
		require.NoError(t, err)
		_, found := conf.FindByID("input:cpu:cpu.1")
		require.False(t, found)
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
		require.NoError(t, err)
		plugin, found := conf.FindByID("input:cpu:cpu.2")
		require.True(t, found)
		require.Equal(t, Plugin{
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
	require.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			require.NoError(t, err)

			var classicConf Config
			err = classicConf.UnmarshalClassic(classicText)
			require.NoError(t, err)
			classicConf.Validate()
		})
	}
}

func TestConfig_Validate_YAML(t *testing.T) {
	names, err := filepath.Glob("testdata/*.yaml")
	require.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			yamlText, err := os.ReadFile(name)
			require.NoError(t, err)

			var yamlConfig Config
			err = yaml.Unmarshal(yamlText, &yamlConfig)
			require.NoError(t, err)
			yamlConfig.Validate()
		})
	}
}

func TestConfig_Validate_JSON(t *testing.T) {
	names, err := filepath.Glob("testdata/*.json")
	require.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			jsonText, err := os.ReadFile(name)
			require.NoError(t, err)

			var jsonConfig Config
			err = json.Unmarshal(jsonText, &jsonConfig)
			require.NoError(t, err)
			jsonConfig.Validate()
		})
	}
}

func Example_configWithProcessors() {
	conf := Config{
		Pipeline: Pipeline{
			Inputs: Plugins{
				Plugin{
					Properties: property.Properties{
						{Key: "name", Value: "dummy"},
						{Key: "processors", Value: property.Properties{
							{Key: "logs", Value: Plugins{
								Plugin{
									Properties: property.Properties{
										{Key: "name", Value: "calyptia"},
										{Key: "actions", Value: property.Properties{
											{Key: "type", Value: "block_keys"},
											{Key: "opts", Value: property.Properties{
												{Key: "regex", Value: "star"},
												{Key: "regexEngine", Value: "pcre2"},
											}},
										}},
									},
								},
							}},
						}},
					},
				},
			},
		},
	}
	yaml, err := conf.DumpAsYAML()
	if err != nil {
		panic(err)
	}

	fmt.Println(yaml)
	// Output:
	// pipeline:
	//     inputs:
	//         - name: dummy
	//           processors:
	//             logs:
	//                 - name: calyptia
	//                   actions:
	//                     type: block_keys
	//                     opts:
	//                         regex: star
	//                         regexEngine: pcre2
}
