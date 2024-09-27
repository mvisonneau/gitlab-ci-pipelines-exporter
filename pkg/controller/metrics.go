package controller

import (
	"context"
	"fmt"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

// Registry wraps a pointer of prometheus.Registry.
type Registry struct {
	*prometheus.Registry

	InternalCollectors struct {
		CurrentlyQueuedTasksCount  prometheus.Collector
		EnvironmentsCount          prometheus.Collector
		ExecutedTasksCount         prometheus.Collector
		GitLabAPIRequestsCount     prometheus.Collector
		GitlabAPIRequestsRemaining prometheus.Collector
		GitlabAPIRequestsLimit     prometheus.Collector
		MetricsCount               prometheus.Collector
		ProjectsCount              prometheus.Collector
		RefsCount                  prometheus.Collector
	}

	Collectors RegistryCollectors
}

// RegistryCollectors ..
type RegistryCollectors map[schemas.MetricKind]prometheus.Collector

// NewRegistry initialize a new registry.
func NewRegistry(ctx context.Context) *Registry {
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
			schemas.MetricKindJobQueuedDurationSeconds:             NewCollectorJobQueuedDurationSeconds(),
			schemas.MetricKindJobRunCount:                          NewCollectorJobRunCount(),
			schemas.MetricKindJobStatus:                            NewCollectorJobStatus(),
			schemas.MetricKindJobTimestamp:                         NewCollectorJobTimestamp(),
			schemas.MetricKindQueuedDurationSeconds:                NewCollectorQueuedDurationSeconds(),
			schemas.MetricKindRunCount:                             NewCollectorRunCount(),
			schemas.MetricKindStatus:                               NewCollectorStatus(),
			schemas.MetricKindTimestamp:                            NewCollectorTimestamp(),
			schemas.MetricKindTestReportTotalTime:                  NewCollectorTestReportTotalTime(),
			schemas.MetricKindTestReportTotalCount:                 NewCollectorTestReportTotalCount(),
			schemas.MetricKindTestReportSuccessCount:               NewCollectorTestReportSuccessCount(),
			schemas.MetricKindTestReportFailedCount:                NewCollectorTestReportFailedCount(),
			schemas.MetricKindTestReportSkippedCount:               NewCollectorTestReportSkippedCount(),
			schemas.MetricKindTestReportErrorCount:                 NewCollectorTestReportErrorCount(),
			schemas.MetricKindTestSuiteTotalTime:                   NewCollectorTestSuiteTotalTime(),
			schemas.MetricKindTestSuiteTotalCount:                  NewCollectorTestSuiteTotalCount(),
			schemas.MetricKindTestSuiteSuccessCount:                NewCollectorTestSuiteSuccessCount(),
			schemas.MetricKindTestSuiteFailedCount:                 NewCollectorTestSuiteFailedCount(),
			schemas.MetricKindTestSuiteSkippedCount:                NewCollectorTestSuiteSkippedCount(),
			schemas.MetricKindTestSuiteErrorCount:                  NewCollectorTestSuiteErrorCount(),
			schemas.MetricKindTestCaseExecutionTime:                NewCollectorTestCaseExecutionTime(),
			schemas.MetricKindTestCaseStatus:                       NewCollectorTestCaseStatus(),
		},
	}

	r.RegisterInternalCollectors()

	if err := r.RegisterCollectors(); err != nil {
		log.WithContext(ctx).
			Fatal(err)
	}

	return r
}

// RegisterInternalCollectors declare our internal collectors to the registry.
func (r *Registry) RegisterInternalCollectors() {
	r.InternalCollectors.CurrentlyQueuedTasksCount = NewInternalCollectorCurrentlyQueuedTasksCount()
	r.InternalCollectors.EnvironmentsCount = NewInternalCollectorEnvironmentsCount()
	r.InternalCollectors.ExecutedTasksCount = NewInternalCollectorExecutedTasksCount()
	r.InternalCollectors.GitLabAPIRequestsCount = NewInternalCollectorGitLabAPIRequestsCount()
	r.InternalCollectors.GitlabAPIRequestsRemaining = NewInternalCollectorGitLabAPIRequestsRemaining()
	r.InternalCollectors.GitlabAPIRequestsLimit = NewInternalCollectorGitLabAPIRequestsLimit()
	r.InternalCollectors.MetricsCount = NewInternalCollectorMetricsCount()
	r.InternalCollectors.ProjectsCount = NewInternalCollectorProjectsCount()
	r.InternalCollectors.RefsCount = NewInternalCollectorRefsCount()

	_ = r.Register(r.InternalCollectors.CurrentlyQueuedTasksCount)
	_ = r.Register(r.InternalCollectors.EnvironmentsCount)
	_ = r.Register(r.InternalCollectors.ExecutedTasksCount)
	_ = r.Register(r.InternalCollectors.GitLabAPIRequestsCount)
	_ = r.Register(r.InternalCollectors.GitlabAPIRequestsRemaining)
	_ = r.Register(r.InternalCollectors.GitlabAPIRequestsLimit)
	_ = r.Register(r.InternalCollectors.MetricsCount)
	_ = r.Register(r.InternalCollectors.ProjectsCount)
	_ = r.Register(r.InternalCollectors.RefsCount)
}

// ExportInternalMetrics ..
func (r *Registry) ExportInternalMetrics(
	ctx context.Context,
	g *gitlab.Client,
	s store.Store,
) (err error) {
	var (
		currentlyQueuedTasks uint64
		environmentsCount    int64
		executedTasksCount   uint64
		metricsCount         int64
		projectsCount        int64
		refsCount            int64
	)

	currentlyQueuedTasks, err = s.CurrentlyQueuedTasksCount(ctx)
	if err != nil {
		return
	}

	executedTasksCount, err = s.ExecutedTasksCount(ctx)
	if err != nil {
		return
	}

	projectsCount, err = s.ProjectsCount(ctx)
	if err != nil {
		return
	}

	environmentsCount, err = s.EnvironmentsCount(ctx)
	if err != nil {
		return
	}

	refsCount, err = s.RefsCount(ctx)
	if err != nil {
		return
	}

	metricsCount, err = s.MetricsCount(ctx)
	if err != nil {
		return
	}

	r.InternalCollectors.CurrentlyQueuedTasksCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(currentlyQueuedTasks))
	r.InternalCollectors.EnvironmentsCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(environmentsCount))
	r.InternalCollectors.ExecutedTasksCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(executedTasksCount))
	r.InternalCollectors.GitLabAPIRequestsCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(g.RequestsCounter.Load()))
	r.InternalCollectors.GitlabAPIRequestsRemaining.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(g.RequestsRemaining))
	r.InternalCollectors.GitlabAPIRequestsLimit.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(g.RequestsLimit))
	r.InternalCollectors.MetricsCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(metricsCount))
	r.InternalCollectors.ProjectsCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(projectsCount))
	r.InternalCollectors.RefsCount.(*prometheus.GaugeVec).With(prometheus.Labels{}).Set(float64(refsCount))

	return
}

// RegisterCollectors add all our metrics to the registry.
func (r *Registry) RegisterCollectors() error {
	for _, c := range r.Collectors {
		if err := r.Register(c); err != nil {
			return fmt.Errorf("could not add provided collector '%v' to the Prometheus registry: %v", c, err)
		}
	}

	return nil
}

// GetCollector ..
func (r *Registry) GetCollector(kind schemas.MetricKind) prometheus.Collector {
	return r.Collectors[kind]
}

// ExportMetrics ..
func (r *Registry) ExportMetrics(metrics schemas.Metrics) {
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
}

func emitStatusMetric(ctx context.Context, s store.Store, metricKind schemas.MetricKind, labelValues map[string]string, statuses []string, status string, sparseMetrics bool) {
	// Moved into separate function to reduce cyclomatic complexity
	// List of available statuses from the API spec
	// ref: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
	for _, currentStatus := range statuses {
		var (
			value        float64
			statusLabels = make(map[string]string)
		)

		for k, v := range labelValues {
			statusLabels[k] = v
		}

		statusLabels["status"] = currentStatus

		statusMetric := schemas.Metric{
			Kind:   metricKind,
			Labels: statusLabels,
			Value:  value,
		}

		if currentStatus == status {
			statusMetric.Value = 1
		} else {
			if sparseMetrics {
				storeDelMetric(ctx, s, statusMetric)

				continue
			}

			statusMetric.Value = 0
		}

		storeSetMetric(ctx, s, statusMetric)
	}
}
