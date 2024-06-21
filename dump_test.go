package fluentbitconfig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/muesli/reflow/dedent"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func TestParseAsYAML(t *testing.T) {
	t.Run("unknown_plural_pipelines", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			pipelines:
				inputs:
					- name: dummy
		`))
		assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: field pipelines not found in type fluentbitconfig.Config")
	})

	t.Run("unknown_section", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			unknown:
				- name: dummy
		`))
		assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: field unknown not found in type fluentbitconfig.Config")
	})

	t.Run("unknown_pipeline_plugins_kind", func(t *testing.T) {
		_, err := ParseAsYAML(configLiteral(`
			pipeline:
				foos:
					- name: bar
		`))
		assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 2: field foos not found in type fluentbitconfig.Pipeline")
	})

	t.Run("empty", func(t *testing.T) {
		got, err := ParseAsYAML("")
		assert.NoError(t, err)
		assert.Equal(t, Config{}, got)
	})

	t.Run("ok", func(t *testing.T) {
		text := configLiteral(`
			pipeline:
				inputs:
					- name: dummy
		`)
		fmt.Printf("text: %q\n", text)
		cfg, err := ParseAsYAML(text)
		assert.NoError(t, err)
		assert.Equal(t, Config{
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

func configLiteral(s string) string {
	if strings.HasPrefix(s, "\n\t") {
		s = dedent.String(s)
	}
	return strings.TrimSpace(strings.ReplaceAll(s, "\t", "    ")) + "\n"
}
