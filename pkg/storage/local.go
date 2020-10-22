package storage

import (
	"sync"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Local ..
type Local struct {
	projects      schemas.Projects
	projectsMutex sync.RWMutex

	projectsRefs      schemas.ProjectsRefs
	projectsRefsMutex sync.RWMutex

	metrics      schemas.Metrics
	metricsMutex sync.RWMutex
}

// SetProject ..
func (l *Local) SetProject(p schemas.Project) error {
	l.projectsMutex.Lock()
	defer l.projectsMutex.Unlock()

	l.projects[p.Key()] = p
	return nil
}

// DelProject ..
func (l *Local) DelProject(k schemas.ProjectKey) error {
	l.projectsMutex.Lock()
	defer l.projectsMutex.Unlock()

	delete(l.projects, k)
	return nil
}

// GetProject ..
func (l *Local) GetProject(p *schemas.Project) error {
	exists, err := l.ProjectExists(p.Key())
	if err != nil {
		return err
	}

	if exists {
		l.projectsMutex.RLock()
		*p = l.projects[p.Key()]
		l.projectsMutex.RUnlock()
	}

	return nil
}

// ProjectExists ..
func (l *Local) ProjectExists(k schemas.ProjectKey) (bool, error) {
	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	_, ok := l.projects[k]
	return ok, nil
}

// Projects ..
func (l *Local) Projects() (projects schemas.Projects, err error) {
	projects = make(schemas.Projects)
	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	for k, v := range l.projects {
		projects[k] = v
	}
	return
}

// ProjectsCount ..
func (l *Local) ProjectsCount() (int64, error) {
	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	return int64(len(l.projects)), nil
}

// SetProjectRef ..
func (l *Local) SetProjectRef(pr schemas.ProjectRef) error {
	l.projectsRefsMutex.Lock()
	defer l.projectsRefsMutex.Unlock()

	l.projectsRefs[pr.Key()] = pr
	return nil
}

// DelProjectRef ..
func (l *Local) DelProjectRef(k schemas.ProjectRefKey) error {
	l.projectsRefsMutex.Lock()
	defer l.projectsRefsMutex.Unlock()

	delete(l.projectsRefs, k)
	return nil
}

// GetProjectRef ..
func (l *Local) GetProjectRef(pr *schemas.ProjectRef) error {
	exists, err := l.ProjectRefExists(pr.Key())
	if err != nil {
		return err
	}

	if exists {
		l.projectsRefsMutex.RLock()
		*pr = l.projectsRefs[pr.Key()]
		l.projectsRefsMutex.RUnlock()
	}

	return nil
}

// ProjectRefExists ..
func (l *Local) ProjectRefExists(k schemas.ProjectRefKey) (bool, error) {
	l.projectsRefsMutex.RLock()
	defer l.projectsRefsMutex.RUnlock()

	_, ok := l.projectsRefs[k]
	return ok, nil
}

// ProjectsRefs ..
func (l *Local) ProjectsRefs() (projectsRefs schemas.ProjectsRefs, err error) {
	projectsRefs = make(schemas.ProjectsRefs)
	l.projectsRefsMutex.RLock()
	defer l.projectsRefsMutex.RUnlock()

	for k, v := range l.projectsRefs {
		projectsRefs[k] = v
	}
	return
}

// ProjectsRefsCount ..
func (l *Local) ProjectsRefsCount() (int64, error) {
	l.projectsRefsMutex.RLock()
	defer l.projectsRefsMutex.RUnlock()

	return int64(len(l.projectsRefs)), nil
}

// SetMetric ..
func (l *Local) SetMetric(m schemas.Metric) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	l.metrics[m.Key()] = m
	return nil
}

// DelMetric ..
func (l *Local) DelMetric(k schemas.MetricKey) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	delete(l.metrics, k)
	return nil
}

// GetMetric ..
func (l *Local) GetMetric(m *schemas.Metric) error {
	exists, err := l.MetricExists(m.Key())
	if err != nil {
		return err
	}

	if exists {
		l.metricsMutex.RLock()
		*m = l.metrics[m.Key()]
		l.metricsMutex.RUnlock()
	}

	return nil
}

// MetricExists ..
func (l *Local) MetricExists(k schemas.MetricKey) (bool, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	_, ok := l.metrics[k]
	return ok, nil
}

// Metrics ..
func (l *Local) Metrics() (metrics schemas.Metrics, err error) {
	metrics = make(schemas.Metrics)
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	for k, v := range l.metrics {
		metrics[k] = v
	}
	return
}

// MetricsCount ..
func (l *Local) MetricsCount() (int64, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	return int64(len(l.metrics)), nil
}
