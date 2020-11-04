package exporter

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Registry wraps a pointer of prometheus.Registry
type Registry struct {
	*prometheus.Registry

	Collectors RegistryCollectors
}

// RegistryCollectors ..
type RegistryCollectors map[schemas.MetricKind]prometheus.Collector

// NewRegistry initialize a new registry
func NewRegistry() *Registry {
	r := &Registry{
		Registry: prometheus.NewRegistry(),
		Collectors: RegistryCollectors{
			schemas.MetricKindCoverage:                             NewCollectorCoverage(),
			schemas.MetricKindDurationSeconds:                      NewCollectorDurationSeconds(),
			schemas.MetricKindEnvironmentBehindCommitsCount:        NewCollectorEnvironmentBehindCommitsCount(),
			schemas.MetricKindEnvironmentBehindDurationSeconds:     NewCollectorEnvironmentBehindDurationSeconds(),
			schemas.MetricKindEnvironmentDeploymentCount:           NewCollectorEnvironmentDeploymentCount(),
			schemas.MetricKindEnvironmentDeploymentDurationSeconds: NewCollectorEnvironmentDeploymentDurationSeconds(),
			schemas.MetricKindEnvironmentDeploymentJobID:           NewCollectorEnvironmentDeploymentJobID(),
			schemas.MetricKindEnvironmentDeploymentStatus:          NewCollectorEnvironmentDeploymentStatus(),
			schemas.MetricKindEnvironmentDeploymentTimestamp:       NewCollectorEnvironmentDeploymentTimestamp(),
			schemas.MetricKindEnvironmentInformation:               NewCollectorEnvironmentInformation(),
			schemas.MetricKindID:                                   NewCollectorID(),
			schemas.MetricKindJobArtifactSizeBytes:                 NewCollectorJobArtifactSizeBytes(),
			schemas.MetricKindJobDurationSeconds:                   NewCollectorJobDurationSeconds(),
			schemas.MetricKindJobID:                                NewCollectorJobID(),
			schemas.MetricKindJobRunCount:                          NewCollectorJobRunCount(),
			schemas.MetricKindJobStatus:                            NewCollectorJobStatus(),
			schemas.MetricKindJobTimestamp:                         NewCollectorJobTimestamp(),
			schemas.MetricKindRunCount:                             NewCollectorRunCount(),
			schemas.MetricKindStatus:                               NewCollectorStatus(),
			schemas.MetricKindTimestamp:                            NewCollectorTimestamp(),
		},
	}

	if err := r.RegisterCollectors(); err != nil {
		log.Fatal(err)
	}

	return r
}

// RegisterCollectors add all our metrics to the registry
func (r *Registry) RegisterCollectors() error {
	for _, c := range r.Collectors {
		if err := r.Register(c); err != nil {
			return fmt.Errorf("could not add provided collector '%v' to the Prometheus registry: %v", c, err)
		}
	}
	return nil
}

// MetricsHandler returns an http handler containing with the desired configuration
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	registry := NewRegistry()

	if err := registry.ExportMetrics(); err != nil {
		log.Error(err.Error())
	}

	promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry:          registry,
		EnableOpenMetrics: config.Server.Metrics.EnableOpenmetricsEncoding,
	}).ServeHTTP(w, r)
}

// GetCollector ..
func (r *Registry) GetCollector(kind schemas.MetricKind) prometheus.Collector {
	return r.Collectors[kind]
}

// ExportMetrics ..
func (r *Registry) ExportMetrics() error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	metrics, err := store.Metrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		switch c := r.GetCollector(m.Kind).(type) {
		case *prometheus.GaugeVec:
			c.With(m.Labels).Set(m.Value)
		case *prometheus.CounterVec:
			c.With(m.Labels).Add(m.Value)
		default:
			log.Errorf("unsupported collector type : %v", reflect.TypeOf(c))
		}
	}

	return nil
}

func emitStatusMetric(metricKind schemas.MetricKind, labelValues map[string]string, statuses []string, status string, sparseMetrics bool) {
	// Moved into separate function to reduce cyclomatic complexity
	// List of available statuses from the API spec
	// ref: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
	for _, s := range statuses {
		var value float64
		statusLabels := make(map[string]string)
		for k, v := range labelValues {
			statusLabels[k] = v
		}
		statusLabels["status"] = s

		statusMetric := schemas.Metric{
			Kind:   metricKind,
			Labels: statusLabels,
			Value:  value,
		}

		if s == status {
			statusMetric.Value = 1
		} else {
			if sparseMetrics {
				storeDelMetric(statusMetric)
				continue
			}
			statusMetric.Value = 0
		}

		storeSetMetric(statusMetric)
	}
}
