package cmd

import (
	"fmt"
	stdlibLog "log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/exporter"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
	taskq.SetLogger(stdlibLog.New(log.StandardLogger().WriterLevel(log.WarnLevel), "taskq", 0))

	// Initialize config
	var cfg schemas.Config
	if cfg, err = schemas.ParseConfigFile(ctx.String("config")); err != nil {
		return
	}

	if len(ctx.String("gitlab-token")) > 0 {
		cfg.Gitlab.Token = ctx.String("gitlab-token")
	}

	if len(cfg.Gitlab.Token) == 0 {
		return fmt.Errorf("--gitlab-token' must be defined")
	}

	if cfg.Server.Webhook.Enabled {
		if len(ctx.String("webhook-secret-token")) > 0 {
			cfg.Server.Webhook.SecretToken = ctx.String("webhook-secret-token")
		}
		if len(cfg.Server.Webhook.SecretToken) == 0 {
			return fmt.Errorf("--webhook-secret-token' must be defined")
		}
	}

	schemas.UpdateProjectDefaults(cfg.ProjectDefaults)

	if len(ctx.String("redis-url")) > 0 {
		cfg.Redis.URL = ctx.String("redis-url")
	}

	if len(cfg.Redis.URL) > 0 {
		log.Info("redis url configured, initializing connection..")
		var opt *redis.Options
		if opt, err = redis.ParseURL(cfg.Redis.URL); err != nil {
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
			"gitlab-endpoint": cfg.Gitlab.URL,
			"pull-rate-limit": fmt.Sprintf("%drps", cfg.Pull.MaximumGitLabAPIRequestsPerSecond),
		},
	).Info("exporter configured")

	log.WithFields(cfg.Pull.ProjectsFromWildcards.Log()).Info("pull projects from wildcards")
	log.WithFields(cfg.Pull.EnvironmentsFromProjects.Log()).Info("pull environments from projects")
	log.WithFields(cfg.Pull.RefsFromProjects.Log()).Info("pull refs from projects")
	log.WithFields(cfg.Pull.Metrics.Log()).Info("pull metrics")
	log.WithFields(cfg.Pull.MetricsWithTraces.Log()).Info("pull metrics with job traces")

	log.WithFields(cfg.GarbageCollect.Projects.Log()).Info("garbage collect projects")
	log.WithFields(cfg.GarbageCollect.Environments.Log()).Info("garbage collect environments")
	log.WithFields(cfg.GarbageCollect.Refs.Log()).Info("garbage collect refs")
	log.WithFields(cfg.GarbageCollect.Metrics.Log()).Info("garbage collect metrics")

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
