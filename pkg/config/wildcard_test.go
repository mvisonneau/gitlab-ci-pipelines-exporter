package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWildcardKey(t *testing.T) {
	w := Wildcard{
		Search: "foo",
	}

	assert.Equal(t, WildcardKey("2203518986"), w.Key())
}
