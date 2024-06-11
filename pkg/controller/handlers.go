package controller

import (
	"context"
	"fmt"
	"io"
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
	span := trace.SpanFromContext(r.Context())
	defer span.End()

	// We create a new background context instead of relying on the request one which has a short cancellation TTL
	ctx := trace.ContextWithSpan(context.Background(), span)

	logger := log.
		WithContext(ctx).
		WithFields(log.Fields{
			"ip-address": r.RemoteAddr,
			"user-agent": r.UserAgent(),
		})

	logger.Debug("webhook request")

	if r.Header.Get("X-Gitlab-Token") != c.Config.Server.Webhook.SecretToken {
		logger.Debug("invalid token provided for a webhook request")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"error\": \"invalid token\"}")

		return
	}

	if r.Body == http.NoBody {
		logger.
			WithError(fmt.Errorf("nil body")).
			Warn("unable to read body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logger.
			WithError(err).
			Warn("unable to read body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	event, err := gitlab.ParseHook(gitlab.HookEventType(r), payload)
	if err != nil {
		logger.
			WithError(err).
			Warn("unable to parse body of a received webhook")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	switch event := event.(type) {
	case *gitlab.PipelineEvent:
		go c.processPipelineEvent(ctx, *event)
	case *gitlab.JobEvent:
		go c.processJobEvent(ctx, *event)
	case *gitlab.DeploymentEvent:
		go c.processDeploymentEvent(ctx, *event)
	case *gitlab.PushEvent:
		go c.processPushEvent(ctx, *event)
	case *gitlab.TagEvent:
		go c.processTagEvent(ctx, *event)
	case *gitlab.MergeEvent:
		go c.processMergeEvent(ctx, *event)
	default:
		logger.
			WithField("event-type", reflect.TypeOf(event).String()).
			Warn("received a non supported event type as a webhook")

		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}
