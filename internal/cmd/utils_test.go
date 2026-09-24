package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
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

func runConfigCliOverrides(t *testing.T, args ...string) config.Config {
	t.Helper()

	cfg := config.New()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "redis-project-ttl"},
			&cli.DurationFlag{Name: "redis-ref-ttl"},
			&cli.DurationFlag{Name: "redis-metric-ttl"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			configCliOverrides(cmd, &cfg)

			return nil
		},
	}

	assert.NoError(t, cmd.Run(context.Background(), append([]string{"gcpe"}, args...)))

	return cfg
}

func TestConfigCliOverridesRedisTTLs(t *testing.T) {
	cfg := runConfigCliOverrides(t,
		"--redis-project-ttl", "720h",
		"--redis-ref-ttl", "720h",
		"--redis-metric-ttl", "720h",
	)

	assert.Equal(t, 720*time.Hour, cfg.Redis.ProjectTTL)
	assert.Equal(t, 720*time.Hour, cfg.Redis.RefTTL)
	assert.Equal(t, 720*time.Hour, cfg.Redis.MetricTTL)
}

func TestConfigCliOverridesRedisTTLsUnset(t *testing.T) {
	cfg := runConfigCliOverrides(t)
	defaults := config.New()

	assert.Equal(t, defaults.Redis.ProjectTTL, cfg.Redis.ProjectTTL)
	assert.Equal(t, defaults.Redis.RefTTL, cfg.Redis.RefTTL)
	assert.Equal(t, defaults.Redis.MetricTTL, cfg.Redis.MetricTTL)
}
