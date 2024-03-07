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

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
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
	var (
		cfg config.Config
		err error
	)

	f, err := ioutil.TempFile(".", "test-*.yml")
	assert.NoError(t, err)

	defer os.Remove(f.Name())

	// Webhook endpoint enabled
	ioutil.WriteFile(f.Name(), []byte(`wildcards: [{}]`), 0o644)

	ctx, flags := NewTestContext()
	flags.String("log-format", "text", "")
	flags.String("log-level", "debug", "")
	flags.String("config", f.Name(), "")

	// Undefined gitlab-token
	flags.String("gitlab-token", "", "")

	_, err = configure(ctx)
	assert.Error(t, err)

	// Valid configuration
	flags.Set("gitlab-token", "secret")

	cfg, err = configure(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "secret", cfg.Gitlab.Token)

	// Invalid config file syntax
	ioutil.WriteFile(f.Name(), []byte("["), 0o644)

	cfg, err = configure(ctx)
	assert.Error(t, err)

	// Webhook endpoint enabled
	ioutil.WriteFile(f.Name(), []byte(`
wildcards: [{}]
server:
  webhook:
    enabled: true
`), 0o644)

	// No secret token defined for the webhook endpoint
	cfg, err = configure(ctx)
	assert.Error(t, err)

	// Defining the webhook secret token
	flags.String("webhook-secret-token", "secret", "")

	cfg, err = configure(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "secret", cfg.Server.Webhook.SecretToken)

	// Test health url flag
	healthURL := "https://gitlab.com/-/readiness?token"
	flags.String("gitlab-health-url", healthURL, "")

	cfg, err = configure(ctx)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Gitlab.HealthURL, healthURL)
	assert.True(t, cfg.Gitlab.EnableHealthCheck)
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
