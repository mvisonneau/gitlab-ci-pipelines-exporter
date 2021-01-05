package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func NewTestContext() (ctx *cli.Context, flags *flag.FlagSet) {
	app := cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"

	app.Metadata = map[string]interface{}{
		"startTime": time.Now(),
	}

	flags = flag.NewFlagSet("test", flag.ContinueOnError)
	ctx = cli.NewContext(app, flags, nil)

	return
}

func TestConfigure(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	ctx, flags := NewTestContext()
	flags.String("log-format", "text", "")
	flags.String("log-level", "debug", "")
	flags.String("config", f.Name(), "")

	// Undefined gitlab-token
	flags.String("gitlab-token", "", "")
	assert.Error(t, configure(ctx))

	// Valid configuration
	flags.Set("gitlab-token", "secret")
	assert.NoError(t, configure(ctx))

	// Invalid config file syntax
	ioutil.WriteFile(f.Name(), []byte("["), 0o644)
	assert.Error(t, configure(ctx))

	// Webhook endpoint enabled
	ioutil.WriteFile(f.Name(), []byte(`
server:
  webhook:
    enabled: true
`), 0o644)

	// No secret token defined for the webhook endpoint
	assert.Error(t, configure(ctx))

	// Defining the webhook secret token
	flags.String("webhook-secret-token", "secret", "")
	assert.NoError(t, configure(ctx))

	// Invalid redis-url
	flags.String("redis-url", "[", "")
	assert.Error(t, configure(ctx))

	// Valid redis-url with unreachable server
	flags.Set("redis-url", "redis://localhost:6379")
	assert.Error(t, configure(ctx))

	// Valid redis-url with reachable server
	// TODO: Figure out how to make it work without failing other tests by timing out
	// s, err := miniredis.Run()
	// if err != nil {
	// 	panic(err)
	// }
	// defer s.Close()

	// flags.Set("redis-url", fmt.Sprintf("redis://%s", s.Addr()))
	// assert.NoError(t, configure(ctx))
}

func TestExit(t *testing.T) {
	err := exit(20, fmt.Errorf("test"))
	assert.Equal(t, "", err.Error())
	assert.Equal(t, 20, err.ExitCode())
}

func TestExecWrapper(t *testing.T) {
	function := func(ctx *cli.Context) (int, error) {
		return 0, nil
	}
	assert.Equal(t, exit(function(&cli.Context{})), ExecWrapper(function)(&cli.Context{}))
}
