package fluentbitconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/muesli/reflow/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func TestParseAsYAML(t *testing.T) {
	t.Run("unknown_plural_pipelines", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			pipelines:
				inputs:
					- name: dummy
		`))
		require.EqualError(t, err, "[1:1] unknown field \"pipelines\"\n>  1 | pipelines:\n       ^\n   2 |     inputs:\n   3 |         - name: dummy")
	})

	t.Run("unknown_section", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			unknown:
				- name: dummy
		`))
		require.EqualError(t, err, "[1:1] unknown field \"unknown\"\n>  1 | unknown:\n       ^\n   2 |     - name: dummy")
	})

	t.Run("unknown_pipeline_plugins_kind", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			pipeline:
				foos:
					- name: bar
		`))
		require.EqualError(t, err, "[2:5] unknown field \"foos\"\n   1 | pipeline:\n>  2 |     foos:\n           ^\n   3 |         - name: bar")
	})

	t.Run("empty", func(t *testing.T) {
		got, err := ParseAsYAML("")
		require.NoError(t, err)
		require.Equal(t, Config{}, got)
	})

	t.Run("ok", func(t *testing.T) {
		text := configLiteral(`
			pipeline:
				inputs:
					- name: dummy
		`)
		fmt.Printf("text: %q\n", text)
		cfg, err := ParseAsYAML(text)
		require.NoError(t, err)
		require.Equal(t, Config{
			Pipeline: Pipeline{
				Inputs: []Plugin{
					{
						ID:   "dummy.0",
						Name: "dummy",
						Properties: property.Properties{
							{
								Key:   "name",
								Value: "dummy",
							},
						},
					},
				},
			},
		}, cfg)
	})
}

func TestMarshaling(t *testing.T) {
	testCases := []struct {
		name          string
		config        Config
		expectClassic string
		expectYAML    string
		expectJSON    string
	}{
		{
			name:          "empty",
			config:        Config{},
			expectClassic: "",
			expectYAML:    `{}`,
			expectJSON: `
{
  "pipeline": {}
}
`,
		},
		{
			name: "partial",
			config: Config{
				Customs: []Plugin{
					{
						ID:   "test-id",
						Name: "test-name",
						Properties: []property.Property{
							{Key: "test-key", Value: "test-value"},
						},
					},
				},
				Pipeline: Pipeline{
					Inputs: []Plugin{
						{Name: "dummy"},
						{},
					},
					Outputs: []Plugin{
						{},
					},
				},
			},
			expectClassic: `
[CUSTOM]
    Name     test-name
    test-key test-value
[INPUT]
    Name dummy
`,
			expectYAML: `
customs:
    - name: test-name
      test-key: test-value
pipeline:
    inputs:
        - name: dummy
        - null
    outputs:
        - null
`,
			expectJSON: `
{
  "customs": [
    {
      "name": "test-name",
      "test-key": "test-value"
    }
  ],
  "pipeline": {
    "inputs": [
      {
        "name": "dummy"
      },
      null
    ],
    "outputs": [
      null
    ]
  }
}
`,
		},
		{
			name: "full",
			config: Config{
				Env: []property.Property{
					{Key: "Apple", Value: "aa"},
					{Key: "banana", Value: "BB"},
				},
				Includes: []string{"i1", "i2"},
				Service: []property.Property{
					{Key: "s1", Value: "hello world"},
				},
				Customs: []Plugin{
					{
						Name: "my custom plugin",
						Properties: []property.Property{
							{Key: "this is a key", Value: "this is a value"},
						},
					},
				},
				Pipeline: Pipeline{
					Inputs: []Plugin{
						{
							Name: "input 1",
							Properties: []property.Property{
								{Key: "a", Value: "a a"},
								{Key: "b", Value: "b b"},
							},
						},
						{
							Name: "input 2",
							Properties: []property.Property{
								{Key: "c", Value: "CC"},
							},
						},
					},
					Filters: []Plugin{
						{
							Name: "filter-1",
							Properties: []property.Property{
								{Key: "e", Value: "EE"},
							},
						},
					},
					Outputs: []Plugin{
						{
							Name: "output-1",
							Properties: []property.Property{
								{Key: "f", Value: "ff \"with a quote\""},
							},
						},
					},
				},
				Parsers: []Plugin{
					{
						Name: "my-parser",
						Properties: []property.Property{
							{Key: "g", Value: "GG"},
						},
					},
				},
			},
			expectClassic: `
@SET Apple=aa
@SET banana=BB
@INCLUDE i1
@INCLUDE i2
[SERVICE]
    s1 hello world
[CUSTOM]
    Name          my custom plugin
    this is a key this is a value
[INPUT]
    Name input 1
    a    a a
    b    b b
[INPUT]
    Name input 2
    c    CC
[FILTER]
    Name filter-1
    e    EE
[OUTPUT]
    Name output-1
    f    ff "with a quote"
[PARSER]
    Name my-parser
    g    GG
`,
			expectYAML: `
env:
    Apple: aa
    banana: BB
includes:
    - i1
    - i2
service:
    s1: hello world
customs:
    - name: my custom plugin
      this is a key: this is a value
pipeline:
    inputs:
        - name: input 1
          a: a a
          b: b b
        - name: input 2
          c: CC
    filters:
        - name: filter-1
          e: EE
    outputs:
        - name: output-1
          f: ff "with a quote"
parsers:
    - name: my-parser
      g: GG
`,
			expectJSON: `
{
  "env": {
    "Apple": "aa",
    "banana": "BB"
  },
  "includes": [
    "i1",
    "i2"
  ],
  "service": {
    "s1": "hello world"
  },
  "customs": [
    {
      "name": "my custom plugin",
      "this is a key": "this is a value"
    }
  ],
  "pipeline": {
    "inputs": [
      {
        "name": "input 1",
        "a": "a a",
        "b": "b b"
      },
      {
        "name": "input 2",
        "c": "CC"
      }
    ],
    "filters": [
      {
        "name": "filter-1",
        "e": "EE"
      }
    ],
    "outputs": [
      {
        "name": "output-1",
        "f": "ff \"with a quote\""
      }
    ]
  },
  "parsers": [
    {
      "name": "my-parser",
      "g": "GG"
    }
  ]
}
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualClassic, err := tc.config.DumpAsClassic()
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tc.expectClassic), strings.TrimSpace(actualClassic))

			actualYAML, err := tc.config.DumpAsYAML()
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tc.expectYAML), strings.TrimSpace(actualYAML))

			// Using a JSON encoder instead of DumpAsJSON for better readability with indented multiline JSON
			var buf bytes.Buffer
			jsonEnc := json.NewEncoder(&buf)
			jsonEnc.SetIndent("", "  ")
			jsonEnc.SetEscapeHTML(false)
			require.NoError(t, jsonEnc.Encode(tc.config))
			assert.Equal(t, strings.TrimSpace(tc.expectJSON), strings.TrimSpace(buf.String()))
		})
	}
}

func configLiteral(s string) string {
	if strings.HasPrefix(s, "\n\t") {
		s = dedent.String(s)
	}
	return strings.TrimSpace(strings.ReplaceAll(s, "\t", "    ")) + "\n"
}
