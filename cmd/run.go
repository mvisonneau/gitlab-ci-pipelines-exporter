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

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", c.Config.Gitlab.URL)
	log.Infof("Polling projects every %ds", c.Config.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %ds", c.Config.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %ds", c.Config.PipelinesPollingIntervalSeconds)
	log.Infof("Global rate limit for the GitLab API set to %d req/s", c.Config.MaximumGitLabAPIRequestsPerSecond)

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	untilStopSignal := make(chan bool)
	pipelinesOnInit := make(chan bool)
	c.OrchestratePolling(untilStopSignal, pipelinesOnInit)
	// get immediately some data from the latest executed pipelines, if configured to do so
	pipelinesOnInit <- c.Config.OnInitFetchRefsFromPipelines

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
	log.Infof("Started listening onto %s", ctx.GlobalString("listen-address"))

	<-onShutdown
	untilStopSignal <- true
	log.Print("Stopped!")

	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxt); err != nil {
		return 1, fmt.Errorf("shutdown failed: %+v", err)
	}

	return 0, nil
}
