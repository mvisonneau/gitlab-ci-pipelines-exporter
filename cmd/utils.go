package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/heptiolabs/healthcheck"
	gcpe "github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/exporter"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/schemas"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/pkg/errors"

	cli "github.com/urfave/cli/v2"
	"github.com/xanzy/go-gitlab"

	log "github.com/sirupsen/logrus"
)

const (
	userAgent = "gitlab-ci-pipelines-exporter"
)

var start time.Time

func configure(ctx *cli.Context) (c *gcpe.Client, h healthcheck.Handler, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	// Configure logger
	lc := &logger.Config{
		Level:  ctx.String("log-level"),
		Format: ctx.String("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return nil, nil, err
	}

	// Initialize config
	cfg := &schemas.Config{}
	if err := cfg.Parse(ctx.String("config")); err != nil {
		return nil, nil, err
	}

	if len(ctx.String("gitlab-token")) > 0 {
		cfg.Gitlab.Token = ctx.String("gitlab-token")
	}

	assertStringVariableDefined(ctx, "listen-address", ctx.String("listen-address"))
	assertStringVariableDefined(ctx, "gitlab-token", cfg.Gitlab.Token)

	// Configure GitLab client
	gitlabHTTPClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.DisableTLSVerify},
	}}

	opts := []gitlab.ClientOptionFunc{
		gitlab.WithHTTPClient(gitlabHTTPClient),
		gitlab.WithBaseURL(cfg.Gitlab.URL),
		gitlab.WithoutRetries(),
	}

	gc, err := gitlab.NewClient(cfg.Gitlab.Token, opts...)
	if err != nil {
		return nil, nil, err
	}

	var redisClient *redis.Client
	var rateLimiter ratelimit.Limiter
	if len(ctx.String("redis-url")) > 0 {
		log.Debug("redis-url flag set, initializing connection..")
		opt, err := redis.ParseURL(ctx.String("redis-url"))
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing redis-url")
		}

		redisClient = redis.NewClient(opt)
		_, err = redisClient.Ping(context.Background()).Result()
		if err != nil {
			return nil, nil, errors.Wrap(err, "connecting to redis")
		}

		rateLimiter = ratelimit.NewRedisLimiter(context.Background(), redisClient, cfg.MaximumGitLabAPIRequestsPerSecond)
	} else {
		rateLimiter = ratelimit.NewLocalLimiter(cfg.MaximumGitLabAPIRequestsPerSecond)
	}

	c = &gcpe.Client{
		Client:      gc,
		RedisClient: redisClient,
		Config:      cfg,
		RateLimiter: rateLimiter,
	}

	c.UserAgent = userAgent

	// Configure liveness and readiness probes
	h = healthcheck.NewHandler()
	if !c.Config.Gitlab.DisableHealthCheck {
		h.AddReadinessCheck("gitlab-reachable", gitlabReadinessCheck(gitlabHTTPClient, c.Config.Gitlab.HealthURL))
	} else {
		log.Warn("GitLab health check has been disabled. Readiness checks won't be operated.")
	}

	return
}

func gitlabReadinessCheck(httpClient *http.Client, url string) healthcheck.Check {
	return func() error {
		client := httpClient
		client.Timeout = 5 * time.Second
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode != 200 {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}
		return err
	}
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
		cli.ShowAppHelp(ctx)
		log.Errorf("'--%s' must be set!", k)
		os.Exit(2)
	}
}
