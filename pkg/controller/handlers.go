package controller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// HealthCheckHandler ..
func (c *Controller) HealthCheckHandler() (h healthcheck.Handler) {
	h = healthcheck.NewHandler()
	if c.Gitlab.EnableHealthCheck {
		h.AddReadinessCheck("gitlab-reachable", c.Gitlab.ReadinessCheck())
	} else {
		log.Warn("GitLab health check has been disabled. Readiness checks won't be operated.")
	}
	return
}

// MetricsHandler ..
func (c *Controller) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	registry := NewRegistry()
	metrics, err := c.Store.Metrics()
	if err != nil {
		log.Error(err.Error())
	}

	registry.ExportMetrics(metrics)
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry:          registry,
		EnableOpenMetrics: c.Server.Metrics.EnableOpenmetricsEncoding,
	}).ServeHTTP(w, r)
}

// WebhookHandler ..
func (c *Controller) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	logFields := log.Fields{
		"ip-address": r.RemoteAddr,
		"user-agent": r.UserAgent(),
	}
	log.WithFields(logFields).Debug("webhook request")

	if r.Header.Get("X-Gitlab-Token") != c.Server.Webhook.SecretToken {
		log.WithFields(logFields).Debug("invalid token provided for a webhook request")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"error\": \"invalid token\"")
		return
	}

	if r.Body == http.NoBody {
		log.WithFields(logFields).WithField("error", "nil body").Warn("unable to read body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Warn("unable to read body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := gitlab.ParseHook(gitlab.HookEventType(r), payload)
	if err != nil {
		log.WithFields(logFields).WithFields(logFields).WithField("error", err.Error()).Warn("unable to parse body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case *gitlab.PipelineEvent:
		go c.processPipelineEvent(*event)
	case *gitlab.DeploymentEvent:
		go c.processDeploymentEvent(*event)
	default:
		log.WithFields(logFields).WithField("event-type", reflect.TypeOf(event).String()).Warn("received a non supported event type as a webhook")
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}
