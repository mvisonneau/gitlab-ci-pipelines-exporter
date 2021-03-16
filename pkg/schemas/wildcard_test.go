package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWildcardKey(t *testing.T) {
	w := Wildcard{
		Search: "foo",
	}

	assert.Equal(t, WildcardKey("1954623779"), w.Key())
}
