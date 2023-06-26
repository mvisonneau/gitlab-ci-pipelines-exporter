package cmd

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/controller"
	monitoringServer "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/server"
)

// Run launches the exporter.
func Run(cliCtx *cli.Context) (int, error) {
	cfg, err := configure(cliCtx)
	if err != nil {
		return 1, err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	c, err := controller.New(ctx, cfg, cliCtx.App.Version)
	if err != nil {
		return 1, err
	}

	// Start the monitoring RPC server
	go func(c *controller.Controller) {
		s := monitoringServer.NewServer(
			c.Gitlab,
			c.Config,
			c.Store,
			c.TaskController.TaskSchedulingMonitoring,
		)
		s.Serve()
	}(&c)

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	// HTTP server
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    cfg.Server.ListenAddress,
		Handler: mux,
	}

	// health endpoints
	health := c.HealthCheckHandler(ctx)
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)

	// metrics endpoint
	if cfg.Server.Metrics.Enabled {
		mux.HandleFunc("/metrics", c.MetricsHandler)
	}

	// pprof/debug endpoints
	if cfg.Server.EnablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// webhook endpoints
	if cfg.Server.Webhook.Enabled {
		mux.HandleFunc("/webhook", c.WebhookHandler)
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithContext(ctx).
				WithError(err).
				Fatal()
		}
	}()

	log.WithFields(
		log.Fields{
			"listen-address":               cfg.Server.ListenAddress,
			"pprof-endpoint-enabled":       cfg.Server.EnablePprof,
			"metrics-endpoint-enabled":     cfg.Server.Metrics.Enabled,
			"webhook-endpoint-enabled":     cfg.Server.Webhook.Enabled,
			"openmetrics-encoding-enabled": cfg.Server.Metrics.EnableOpenmetricsEncoding,
			"controller-uuid":              c.UUID,
		},
	).Info("http server started")

	<-onShutdown
	log.Info("received signal, attempting to gracefully exit..")
	ctxCancel()

	httpServerContext, forceHTTPServerShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer forceHTTPServerShutdown()

	if err := srv.Shutdown(httpServerContext); err != nil {
		return 1, err
	}

	log.Info("stopped!")

	return 0, nil
}
