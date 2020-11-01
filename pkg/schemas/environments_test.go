package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentKey(t *testing.T) {
	e := Environment{
		ProjectName: "foo",
		Name:        "bar",
	}

	assert.Equal(t, EnvironmentKey("2666930069"), e.Key())
}

func TestEnvironmentsCount(t *testing.T) {
	assert.Equal(t, 2, Environments{
		EnvironmentKey("foo"): Environment{},
		EnvironmentKey("bar"): Environment{},
	}.Count())
}
