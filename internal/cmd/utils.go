package cmd

import (
	"fmt"
	stdlibLog "log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/exporter"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var start time.Time

func configure(ctx *cli.Context) (err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	// Configure logger
	if err = logger.Configure(logger.Config{
		Level:  ctx.String("log-level"),
		Format: ctx.String("log-format"),
	}); err != nil {
		return
	}

	// This hack is to embed taskq logs with logrus
	taskq.SetLogger(stdlibLog.New(log.StandardLogger().Writer(), "taskq", 0))

	// Initialize config
	cfg := schemas.Config{}
	if cfg, err = schemas.ParseConfigFile(ctx.String("config")); err != nil {
		return
	}

	if len(ctx.String("gitlab-token")) > 0 {
		cfg.Gitlab.Token = ctx.String("gitlab-token")
	}
	assertStringVariableDefined(ctx, "gitlab-token", cfg.Gitlab.Token)

	if cfg.Server.Webhook.Enabled {
		if len(ctx.String("webhook-secret-token")) > 0 {
			cfg.Server.Webhook.SecretToken = ctx.String("webhook-secret-token")
		}
		assertStringVariableDefined(ctx, "webhook-secret-token", cfg.Server.Webhook.SecretToken)
	}

	schemas.UpdateProjectDefaults(cfg.ProjectDefaults)

	if len(ctx.String("redis-url")) > 0 {
		cfg.Redis.URL = ctx.String("redis-url")
	}

	if len(cfg.Redis.URL) > 0 {
		log.Debug("redis url configured, initializing connection..")
		var opt *redis.Options
		if opt, err = redis.ParseURL(ctx.String("redis-url")); err != nil {
			return errors.Wrap(err, "parsing redis-url")
		}

		if err = exporter.ConfigureRedisClient(redis.NewClient(opt)); err != nil {
			return
		}
	}

	if err = exporter.Configure(cfg, ctx.App.Version); err != nil {
		return
	}

	log.WithFields(
		log.Fields{
			"gitlab-endpoint":                           cfg.Gitlab.URL,
			"pull-projects-from-wildcards-interval":     fmt.Sprintf("%ds", cfg.Pull.ProjectsFromWildcards.IntervalSeconds),
			"pull-projects-refs-from-projects-interval": fmt.Sprintf("%ds", cfg.Pull.ProjectRefsFromProjects.IntervalSeconds),
			"pull-projects-refs-metrics-interval":       fmt.Sprintf("%ds", cfg.Pull.ProjectRefsMetrics.IntervalSeconds),
			"pull-rate-limit":                           fmt.Sprintf("%drps", cfg.Pull.MaximumGitLabAPIRequestsPerSecond),
		},
	).Info("exporter configured")

	return
}

func exit(exitCode int, err error) cli.ExitCoder {
	defer log.WithFields(
		log.Fields{
			"execution-time": time.Since(start),
		},
	).Debug("exited..")

	if err != nil {
		log.Error(err.Error())
	}

	return cli.NewExitError("", exitCode)
}

// ExecWrapper gracefully logs and exits our `run` functions
func ExecWrapper(f func(ctx *cli.Context) (int, error)) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}

func assertStringVariableDefined(ctx *cli.Context, k, v string) {
	if len(v) == 0 {
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("'--%s' must be set!", k)
		os.Exit(2)
	}
}
