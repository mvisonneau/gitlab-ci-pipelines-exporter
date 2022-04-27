package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestRunWrongLogLevel(t *testing.T) {
// 	ctx, flags := NewTestContext()
// 	flags.String("log-format", "foo", "")
// 	exitCode, err := Run(ctx)
// 	assert.Equal(t, 1, exitCode)
// 	assert.Error(t, err)
// }

func TestRunInvalidConfigFile(t *testing.T) {
	ctx, flags := NewTestContext()

	flags.String("config", "path_does_not_exist", "")

	exitCode, err := Run(ctx)
	assert.Equal(t, 1, exitCode)
	assert.Error(t, err)
}
