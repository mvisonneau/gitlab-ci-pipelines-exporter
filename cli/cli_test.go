package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCli(t *testing.T) {
	version := "0.0.0"
	app := Init(&version)
	assert.Equal(t, app.Name, "gitlab-ci-pipelines-exporter")
	assert.Equal(t, app.Version, "0.0.0")
}
