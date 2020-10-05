package storage

import (
	"sync"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Local ..
type Local struct {
	mutex        sync.Mutex
	projects     schemas.Projects
	projectsRefs schemas.ProjectsRefs
	metrics      schemas.Metrics
}

// SetProject ..
func (l *Local) SetProject(p schemas.Project) error {
	l.mutex.Lock()
	l.projects[p.Key()] = p
	l.mutex.Unlock()
	return nil
}

// DelProject ..
func (l *Local) DelProject(k schemas.ProjectKey) error {
	l.mutex.Lock()
	delete(l.projects, k)
	l.mutex.Unlock()
	return nil
}

// ProjectExists ..
func (l *Local) ProjectExists(k schemas.ProjectKey) (bool, error) {
	_, ok := l.projects[k]
	return ok, nil
}

// Projects ..
func (l *Local) Projects() (schemas.Projects, error) {
	return l.projects, nil
}

// ProjectsCount ..
func (l *Local) ProjectsCount() (int64, error) {
	return int64(len(l.projects)), nil
}

// SetProjectRef ..
func (l *Local) SetProjectRef(pr schemas.ProjectRef) error {
	l.mutex.Lock()
	l.projectsRefs[pr.Key()] = pr
	l.mutex.Unlock()
	return nil
}

// DelProjectRef ..
func (l *Local) DelProjectRef(k schemas.ProjectRefKey) error {
	l.mutex.Lock()
	delete(l.projectsRefs, k)
	l.mutex.Unlock()
	return nil
}

// ProjectRefExists ..
func (l *Local) ProjectRefExists(k schemas.ProjectRefKey) (bool, error) {
	_, ok := l.projectsRefs[k]
	return ok, nil
}

// ProjectsRefs ..
func (l *Local) ProjectsRefs() (schemas.ProjectsRefs, error) {
	return l.projectsRefs, nil
}

// ProjectsRefsCount ..
func (l *Local) ProjectsRefsCount() (int64, error) {
	return int64(len(l.projectsRefs)), nil
}

// SetMetric ..
func (l *Local) SetMetric(m schemas.Metric) error {
	l.mutex.Lock()
	l.metrics[m.Key()] = m
	l.mutex.Unlock()
	return nil
}

// DelMetric ..
func (l *Local) DelMetric(k schemas.MetricKey) error {
	l.mutex.Lock()
	delete(l.metrics, k)
	l.mutex.Unlock()
	return nil
}

// MetricExists ..
func (l *Local) MetricExists(k schemas.MetricKey) (bool, error) {
	_, ok := l.metrics[k]
	return ok, nil
}

// PullMetricValue ..
func (l *Local) PullMetricValue(_ *schemas.Metric) error {
	return nil
}

// Metrics ..
func (l *Local) Metrics() (schemas.Metrics, error) {
	return l.metrics, nil
}

// MetricsCount ..
func (l *Local) MetricsCount() (int64, error) {
	return int64(len(l.metrics)), nil
}
