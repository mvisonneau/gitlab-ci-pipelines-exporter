package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectKey(t *testing.T) {
	assert.Equal(t, ProjectKey("2356372769"), NewProject("foo").Key())
}
