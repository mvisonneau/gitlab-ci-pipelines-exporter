package cmd

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/exporter"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/pkg/errors"

	cli "github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

var start time.Time

func configure(ctx *cli.Context) (h healthcheck.Handler, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	// Configure logger
	lc := &logger.Config{
		Level:  ctx.String("log-level"),
		Format: ctx.String("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return nil, err
	}

	// Initialize config
	if err := exporter.Config.Parse(ctx.String("config")); err != nil {
		return nil, err
	}

	if len(ctx.String("gitlab-token")) > 0 {
		exporter.Config.Gitlab.Token = ctx.String("gitlab-token")
	}

	assertStringVariableDefined(ctx, "listen-address", ctx.String("listen-address"))
	assertStringVariableDefined(ctx, "gitlab-token", exporter.Config.Gitlab.Token)

	schemas.UpdateProjectDefaults(exporter.Config.Defaults)

	var rateLimiter ratelimit.Limiter
	if len(ctx.String("redis-url")) > 0 {
		log.Debug("redis-url flag set, initializing connection..")
		opt, err := redis.ParseURL(ctx.String("redis-url"))
		if err != nil {
			return nil, errors.Wrap(err, "parsing redis-url")
		}

		redisClient := redis.NewClient(opt)
		if err := exporter.ConfigureRedisClient(redisClient); err != nil {
			return nil, err
		}

		rateLimiter = ratelimit.NewRedisLimiter(context.Background(), redisClient, exporter.Config.MaximumGitLabAPIRequestsPerSecond)
	} else {
		rateLimiter = ratelimit.NewLocalLimiter(exporter.Config.MaximumGitLabAPIRequestsPerSecond)
	}

	gc, err := gitlab.NewClient(gitlab.ClientConfig{
		URL:              exporter.Config.Gitlab.URL,
		Token:            exporter.Config.Gitlab.Token,
		DisableTLSVerify: exporter.Config.Gitlab.DisableTLSVerify,
		UserAgentVersion: ctx.App.Version,
		RateLimiter:      rateLimiter,
		ReadinessURL:     exporter.Config.Gitlab.HealthURL,
	})

	if err != nil {
		return nil, err
	}

	exporter.ConfigureGitlabClient(gc)
	exporter.ConfigurePollingQueue()
	exporter.ConfigureStore()

	h = healthcheck.NewHandler()
	if !exporter.Config.Gitlab.DisableHealthCheck {
		h.AddReadinessCheck("gitlab-reachable", gc.ReadinessCheck())
	} else {
		log.Warn("GitLab health check has been disabled. Readiness checks won't be operated.")
	}

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
