package fluentbitconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	fluentbitconfig "github.com/calyptia/go-fluentbit-config"
	"github.com/calyptia/go-fluentbit-config/classic"
	"gopkg.in/yaml.v3"
)

func Test_Config(t *testing.T) {
	names, err := filepath.Glob("testdata/*.conf")
	assert.NoError(t, err)

	for _, name := range names {
		t.Run(makeTestName(name), func(t *testing.T) {
			classicText, err := os.ReadFile(name)
			assert.NoError(t, err)

			var classicConf classic.Classic
			err = classicConf.UnmarshalText(classicText)
			assert.NoError(t, err)

			yamlText, err := os.ReadFile(strings.Replace(name, ".conf", ".yaml", 1))
			assert.NoError(t, err)

			var yamlConf fluentbitconfig.Config
			err = yaml.Unmarshal(yamlText, &yamlConf)
			assert.NoError(t, err)

			gotYamlText, err := yaml.Marshal(classicConf.ToConfig())
			assert.NoError(t, err)
			assert.Equal(t, string(yamlText), string(gotYamlText))

			gotClassicText, err := classic.FromConfig(yamlConf).MarshalText()
			assert.NoError(t, err)
			assert.Equal(t, string(classicText), string(gotClassicText))
		})
	}
}

func unique[T comparable](s []T) []T {
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

func filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

func makeTestName(fp string) string {
	s := filepath.Base(fp)
	s = strings.TrimRight(s, filepath.Ext(s))
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ReplaceAll(s, " ", "_")
}
