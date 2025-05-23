package property

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProperties_Equal(t *testing.T) {
	empty := Properties{}
	p1 := Properties{
		{Key: "key1", Value: []any{"value1"}},
	}
	p1Copy := Properties{
		{Key: "key1", Value: []any{"value1"}},
	}
	p2 := Properties{
		{Key: "key1", Value: []any{"value1"}},
		{Key: "key2", Value: []any{"value2"}},
	}

	require.True(t, p1.Equal(p1))
	require.True(t, p1.Equal(p1Copy))
	require.True(t, empty.Equal(empty))
	require.True(t, empty.Equal(nil))
	require.True(t, ((*Properties)(nil)).Equal(empty))

	require.False(t, p1.Equal(p2))
	require.False(t, p2.Equal(p1))
}
