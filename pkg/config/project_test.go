package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectKey(t *testing.T) {
	p := Project{
		Name: "foo",
	}

	assert.Equal(t, ProjectKey("2356372769"), p.Key())
}
