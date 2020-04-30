package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
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
			"gitlab-endpoint":                     c.Config.Gitlab.URL,
			"discover-wildcard-projects-interval": fmt.Sprintf("%ds", c.Config.WildcardsProjectsDiscoverIntervalSeconds),
			"discover-projects-refs-interval":     fmt.Sprintf("%ds", c.Config.ProjectsRefsDiscoverIntervalSeconds),
			"polling-projects-refs-interval":      fmt.Sprintf("%ds", c.Config.ProjectsRefsPollingIntervalSeconds),
			"rate-limit":                          fmt.Sprintf("%drps", c.Config.MaximumGitLabAPIRequestsPerSecond),
			"polling-workers":                     c.Config.PollingWorkers,
			"on-init-fetch-refs-from-pipelines":   c.Config.OnInitFetchRefsFromPipelines,
		},
	).Info("starting exporter")

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	// Register the default metrics into a new registry
	registry := gcpe.NewRegistry()
	if err := registry.RegisterDefaultMetrics(); err != nil {
		return 1, err
	}

	orchestratePollingContext, stopOrchestratePolling := context.WithCancel(context.Background())
	c.OrchestratePolling(orchestratePollingContext)

	// Expose the registered registry via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", registry.MetricsHandler(c.Config.DisableOpenmetricsEncoding))

	if ctx.GlobalBool("enable-pprof") {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

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
	log.Info("received signal, attempting to gracefully exit..")
	stopOrchestratePolling()

	httpServerContext, forceHTTPServerShutdown := context.WithTimeout(orchestratePollingContext, 5*time.Second)
	defer forceHTTPServerShutdown()

	if err := srv.Shutdown(httpServerContext); err != nil {
		return 1, fmt.Errorf("shutdown failed: %+v", err)
	}

	log.Info("stopped!")
	return 0, nil
}
