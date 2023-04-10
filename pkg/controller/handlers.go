package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// HealthCheckHandler ..
func (c *Controller) HealthCheckHandler(ctx context.Context) (h healthcheck.Handler) {
	h = healthcheck.NewHandler()
	if c.Config.Gitlab.EnableHealthCheck {
		h.AddReadinessCheck("gitlab-reachable", c.Gitlab.ReadinessCheck(ctx))
	} else {
		log.WithContext(ctx).
			Warn("GitLab health check has been disabled. Readiness checks won't be operated.")
	}

	return
}

// MetricsHandler ..
func (c *Controller) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	defer span.End()

	registry := NewRegistry(ctx)

	metrics, err := c.Store.Metrics(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	if err := registry.ExportInternalMetrics(
		ctx,
		c.Gitlab,
		c.Store,
	); err != nil {
		log.WithContext(ctx).
			WithError(err).
			Warn()
	}

	registry.ExportMetrics(metrics)

	otelhttp.NewHandler(
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			Registry:          registry,
			EnableOpenMetrics: c.Config.Server.Metrics.EnableOpenmetricsEncoding,
		}),
		"/metrics",
	).ServeHTTP(w, r)
}

// WebhookHandler ..
func (c *Controller) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	defer span.End()

	logFields := log.Fields{
		"ip-address": r.RemoteAddr,
		"user-agent": r.UserAgent(),
	}

	log.WithFields(logFields).Debug("webhook request")

	if r.Header.Get("X-Gitlab-Token") != c.Config.Server.Webhook.SecretToken {
		log.WithFields(logFields).Debug("invalid token provided for a webhook request")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"error\": \"invalid token\"}")

		return
	}

	if r.Body == http.NoBody {
		log.WithContext(ctx).
			WithFields(logFields).
			WithError(fmt.Errorf("nil body")).
			Warn("unable to read body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			WithError(err).
			Warn("unable to read body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	event, err := gitlab.ParseHook(gitlab.HookEventType(r), payload)
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			WithError(err).
			Warn("unable to parse body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	switch event := event.(type) {
	case *gitlab.PipelineEvent:
		go c.processPipelineEvent(*event)
	case *gitlab.DeploymentEvent:
		go c.processDeploymentEvent(*event)
	default:
		log.WithContext(ctx).
			WithFields(logFields).
			WithField("event-type", reflect.TypeOf(event).String()).
			Warn("received a non supported event type as a webhook")

		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}
