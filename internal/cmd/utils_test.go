package cmd

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	cli "github.com/urfave/cli/v2"
)

func NewTestContext() (ctx *cli.Context, flags *flag.FlagSet) {
	app := cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"

	app.Metadata = map[string]interface{}{
		"startTime": time.Now(),
	}

	flags = flag.NewFlagSet("test", flag.ContinueOnError)
	ctx = cli.NewContext(app, flags, nil)

	flags.String("log-level", "fatal", "")

	return
}

func TestExit(t *testing.T) {
	err := exit(20, fmt.Errorf("test"))
	assert.Equal(t, "", err.Error())
	assert.Equal(t, 20, err.ExitCode())
}
