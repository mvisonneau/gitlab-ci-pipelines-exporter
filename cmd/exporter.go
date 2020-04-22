package cmd

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

const (
	userAgent = "gitlab-ci-pipelines-exporter"
)

// Run launches the exporter
func Run(ctx *cli.Context) error {

	// Configure logger
	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return exit(err, 1)
	}

	// Initialize config
	if err := cfg.Parse(ctx.GlobalString("config")); err != nil {
		return exit(err, 1)
	}

	cfg.MergeWithContext(ctx)

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", cfg.Gitlab.URL)
	log.Infof("Polling projects every %ds", cfg.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %ds", cfg.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %ds", cfg.PipelinesPollingIntervalSeconds)
	log.Infof("Global rate limit for the GitLab API set to %d req/s", cfg.MaximumGitLabAPIRequestsPerSecond)

	// Configure GitLab client
	opts := []gitlab.ClientOptionFunc{
		gitlab.WithHTTPClient(&http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.SkipTLSVerify},
		}}),
		gitlab.WithBaseURL(cfg.Gitlab.URL),
	}
	gc, err := gitlab.NewClient(cfg.Gitlab.Token, opts...)
	if err != nil {
		return exit(err, 1)
	}
	c := &Client{
		Client:      gc,
		RateLimiter: ratelimit.New(cfg.MaximumGitLabAPIRequestsPerSecond),
	}
	c.UserAgent = userAgent

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	untilStopSignal := make(chan bool)
	c.pollProjects(untilStopSignal)

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	if !cfg.Gitlab.SkipHealthCheck {
		health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(cfg.Gitlab.HealthURL, 5*time.Second))
	} else {
		log.Warn("TLS verification has been disabled. Readiness checks won't be operated.")
	}

	// Register the default metrics into a new registry
	registerMetricOn(registry, log.StandardLogger(), defaultMetrics...)

	// Expose the registered registry via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", metricsHandlerFor(registry, cfg.PrometheusOpenmetricsEncoding))


	srv := &http.Server{
		Addr:    ctx.GlobalString("listen-address"),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Infof("Started listening onto %s", ctx.GlobalString("listen-address"))

	<-onShutdown
	untilStopSignal <- true
	log.Print("Stopped!")

	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxt); err != nil {
		log.Fatalf("Shutdown failed: %+v", err)
	}

	return exit(nil, 0)
}

func metricsHandlerFor(registry *prometheus.Registry, openMetricsEncoder bool) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry:          registry,
		EnableOpenMetrics: openMetricsEncoder,
	})
}

func exit(err error, exitCode int) *cli.ExitError {
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError(err.Error(), exitCode)
	}

	return cli.NewExitError("", exitCode)
}
