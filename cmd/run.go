package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gcpe "github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/exporter"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Run launches the exporter
func Run(ctx *cli.Context) (int, error) {
	c, health, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	log.WithFields(
		log.Fields{
			"gitlab-endpoint":                   c.Config.Gitlab.URL,
			"polling-projects-every":            fmt.Sprintf("%ds", c.Config.ProjectsPollingIntervalSeconds),
			"polling-refs-every":                fmt.Sprintf("%ds", c.Config.RefsPollingIntervalSeconds),
			"polling-pipelines-every":           fmt.Sprintf("%ds", c.Config.PipelinesPollingIntervalSeconds),
			"rate-limit":                        fmt.Sprintf("%drps", c.Config.MaximumGitLabAPIRequestsPerSecond),
			"on-init-fetch-refs-from-pipelines": c.Config.OnInitFetchRefsFromPipelines,
		},
	).Info("starting exporter")

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	untilStopSignal := make(chan bool)
	c.OrchestratePolling(untilStopSignal, c.Config.OnInitFetchRefsFromPipelines)

	// Register the default metrics into a new registry
	registry := gcpe.NewRegistry()
	if err := registry.RegisterDefaultMetrics(); err != nil {
		return 1, err
	}

	// Expose the registered registry via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", registry.MetricsHandler(c.Config.DisableOpenmetricsEncoding))

	srv := &http.Server{
		Addr:    ctx.GlobalString("listen-address"),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.WithFields(
		log.Fields{
			"listen-address": ctx.GlobalString("listen-address"),
		},
	).Info("started, now serving requests")

	<-onShutdown
	untilStopSignal <- true
	log.Info("stopped!")

	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxt); err != nil {
		return 1, fmt.Errorf("shutdown failed: %+v", err)
	}

	return 0, nil
}
