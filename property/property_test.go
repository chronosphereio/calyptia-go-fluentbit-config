package property

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestProperties_Equal(t *testing.T) {
	assert.NotPanics(t, func() {
		p1 := Properties{
			{Key: "key1", Value: []any{"value1"}},
		}
		p2 := Properties{
			{Key: "key1", Value: []any{"value1"}},
		}
		p1.Equal(p2)
	})
}
