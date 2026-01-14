package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestExit(t *testing.T) {
	err := exit(20, fmt.Errorf("test"))
	assert.Equal(t, "", err.Error())
	assert.Equal(t, 20, err.ExitCode())
}

func TestExecWrapper(t *testing.T) {
	function := func(_ context.Context, _ *cli.Command) (int, error) {
		return 0, nil
	}

	err := ExecWrapper(function)(context.Background(), &cli.Command{})
	exitErr, ok := err.(cli.ExitCoder)
	assert.True(t, ok)
	assert.Equal(t, 0, exitErr.ExitCode())
}
