package fluentbitconfig

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func TestPlugins_IDs(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		conf, err := ParseAs(`
			[INPUT]
				name cpu
			[INPUT]
				name mem
			[INPUT]
				name  cpu
		`, FormatClassic)
		require.NoError(t, err)
		require.Equal(t, []string{
			"cpu.0",
			"mem.1",
			"cpu.2",
		}, conf.Pipeline.Inputs.IDs())
	})
}

func TestPlugins_FindByID(t *testing.T) {
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
		_, found := conf.Pipeline.Inputs.FindByID("cpu.1")
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
		plugin, found := conf.Pipeline.Inputs.FindByID("cpu.2")
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
