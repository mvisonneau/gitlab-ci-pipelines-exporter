package storage

import (
	"sync"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Local ..
type Local struct {
	projects      schemas.Projects
	projectsMutex sync.RWMutex

	environments      schemas.Environments
	environmentsMutex sync.RWMutex

	refs      schemas.Refs
	refsMutex sync.RWMutex

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

// SetEnvironment ..
func (l *Local) SetEnvironment(environment schemas.Environment) error {
	l.environmentsMutex.Lock()
	defer l.environmentsMutex.Unlock()

	l.environments[environment.Key()] = environment
	return nil
}

// DelEnvironment ..
func (l *Local) DelEnvironment(k schemas.EnvironmentKey) error {
	l.environmentsMutex.Lock()
	defer l.environmentsMutex.Unlock()

	delete(l.environments, k)
	return nil
}

// GetEnvironment ..
func (l *Local) GetEnvironment(environment *schemas.Environment) error {
	exists, err := l.EnvironmentExists(environment.Key())
	if err != nil {
		return err
	}

	if exists {
		l.environmentsMutex.RLock()
		*environment = l.environments[environment.Key()]
		l.environmentsMutex.RUnlock()
	}

	return nil
}

// EnvironmentExists ..
func (l *Local) EnvironmentExists(k schemas.EnvironmentKey) (bool, error) {
	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	_, ok := l.environments[k]
	return ok, nil
}

// Environments ..
func (l *Local) Environments() (environments schemas.Environments, err error) {
	environments = make(schemas.Environments)
	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	for k, v := range l.environments {
		environments[k] = v
	}
	return
}

// EnvironmentsCount ..
func (l *Local) EnvironmentsCount() (int64, error) {
	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	return int64(len(l.environments)), nil
}

// SetRef ..
func (l *Local) SetRef(ref schemas.Ref) error {
	l.refsMutex.Lock()
	defer l.refsMutex.Unlock()

	l.refs[ref.Key()] = ref
	return nil
}

// DelRef ..
func (l *Local) DelRef(k schemas.RefKey) error {
	l.refsMutex.Lock()
	defer l.refsMutex.Unlock()

	delete(l.refs, k)
	return nil
}

// GetRef ..
func (l *Local) GetRef(ref *schemas.Ref) error {
	exists, err := l.RefExists(ref.Key())
	if err != nil {
		return err
	}

	if exists {
		l.refsMutex.RLock()
		*ref = l.refs[ref.Key()]
		l.refsMutex.RUnlock()
	}

	return nil
}

// RefExists ..
func (l *Local) RefExists(k schemas.RefKey) (bool, error) {
	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	_, ok := l.refs[k]
	return ok, nil
}

// Refs ..
func (l *Local) Refs() (refs schemas.Refs, err error) {
	refs = make(schemas.Refs)
	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	for k, v := range l.refs {
		refs[k] = v
	}
	return
}

// RefsCount ..
func (l *Local) RefsCount() (int64, error) {
	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	return int64(len(l.refs)), nil
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
