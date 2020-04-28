package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"

	gcpe "github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/exporter"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/schemas"

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
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return nil, nil, err
	}

	// Initialize config
	cfg := &schemas.Config{}
	if err := cfg.Parse(ctx.GlobalString("config")); err != nil {
		return nil, nil, err
	}

	if len(ctx.GlobalString("gitlab-token")) > 0 {
		cfg.Gitlab.Token = ctx.GlobalString("gitlab-token")
	}

	// Configure GitLab client
	gitlabHTTPClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.DisableTLSVerify},
	}}

	opts := []gitlab.ClientOptionFunc{
		gitlab.WithHTTPClient(gitlabHTTPClient),
		gitlab.WithBaseURL(cfg.Gitlab.URL),
	}

	gc, err := gitlab.NewClient(cfg.Gitlab.Token, opts...)
	if err != nil {
		return nil, nil, err
	}

	c = &gcpe.Client{
		Client:      gc,
		Config:      cfg,
		RateLimiter: ratelimit.New(cfg.MaximumGitLabAPIRequestsPerSecond),
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

func exit(exitCode int, err error) *cli.ExitError {
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
func ExecWrapper(f func(ctx *cli.Context) (int, error)) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}
