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

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/exporter"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// Run launches the exporter
func Run(ctx *cli.Context) (int, error) {
	health, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	log.WithFields(
		log.Fields{
			"gitlab-endpoint":                     exporter.Config.Gitlab.URL,
			"discover-wildcard-projects-interval": fmt.Sprintf("%ds", exporter.Config.WildcardsProjectsDiscoverIntervalSeconds),
			"discover-projects-refs-interval":     fmt.Sprintf("%ds", exporter.Config.ProjectsRefsDiscoverIntervalSeconds),
			"polling-projects-refs-interval":      fmt.Sprintf("%ds", exporter.Config.ProjectsRefsPollingIntervalSeconds),
			"rate-limit":                          fmt.Sprintf("%drps", exporter.Config.MaximumGitLabAPIRequestsPerSecond),
			"on-init-fetch-refs-from-pipelines":   exporter.Config.OnInitFetchRefsFromPipelines,
		},
	).Info("starting exporter")

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	schedulingContext, stopOrchestratePolling := context.WithCancel(context.Background())
	exporter.Schedule(schedulingContext)
	exporter.ProcessPollingQueue(schedulingContext)

	// Expose the registered registry via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.HandleFunc("/metrics", exporter.MetricsHandler)

	if ctx.Bool("enable-pprof") {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	srv := &http.Server{
		Addr:    ctx.String("listen-address"),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.WithFields(
		log.Fields{
			"listen-address": ctx.String("listen-address"),
		},
	).Info("started, now serving requests")

	<-onShutdown
	log.Info("received signal, attempting to gracefully exit..")
	stopOrchestratePolling()

	httpServerContext, forceHTTPServerShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer forceHTTPServerShutdown()

	if err := srv.Shutdown(httpServerContext); err != nil {
		log.Fatalf("shutdown failed: %+v", err)
	}

	log.Info("stopped!")
	return 0, nil
}
