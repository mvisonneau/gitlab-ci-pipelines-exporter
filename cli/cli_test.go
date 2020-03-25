package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCli(t *testing.T) {
	version := "0.0.0"
	app := Init(&version)
	assert.Equal(t, "gitlab-ci-pipelines-exporter", app.Name)
	assert.Equal(t, "0.0.0", app.Version)
}
