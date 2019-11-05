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
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
)

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
}

var (
	coverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		[]string{"project", "ref"},
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"project", "ref"},
	)

	lastRunID = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		[]string{"project", "ref"},
	)

	lastRunStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		[]string{"project", "ref", "status"},
	)

	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		[]string{"project", "ref"},
	)

	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"project", "ref"},
	)
)

func init() {
	cfg = &Config{}
	prometheus.MustRegister(coverage)
	prometheus.MustRegister(lastRunDuration)
	prometheus.MustRegister(lastRunID)
	prometheus.MustRegister(lastRunStatus)
	prometheus.MustRegister(runCount)
	prometheus.MustRegister(timeSinceLastRun)
}

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

	// Parse config file
	if err := cfg.Parse(ctx.GlobalString("config")); err != nil {
		return exit(err, 1)
	}

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", cfg.Gitlab.URL)
	log.Infof("Polling projects every %vs", cfg.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %vs", cfg.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %vs", cfg.PipelinesPollingIntervalSeconds)

	// Configure GitLab client
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.SkipTLSVerify},
	}

	c := &Client{
		gitlab.NewClient(&http.Client{Transport: httpTransport}, cfg.Gitlab.Token),
	}
	c.SetBaseURL(cfg.Gitlab.URL)

	go c.pollProjects()

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	if !cfg.Gitlab.SkipTLSVerify {
		health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(cfg.Gitlab.HealthURL, 5*time.Second))
	} else {
		log.Warn("TLS verification has been disabled. Readiness checks won't be operated.")
	}

	// Graceful shutdowns
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Expose the registered metrics via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", promhttp.Handler())

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

	<-done
	log.Print("Stopped!")

	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxt); err != nil {
		log.Fatalf("Shutdown failed: %+v", err)
	}

	return exit(nil, 0)
}

func exit(err error, exitCode int) *cli.ExitError {
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError(err.Error(), exitCode)
	}

	return cli.NewExitError("", exitCode)
}
