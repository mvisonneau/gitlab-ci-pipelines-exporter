package store

import (
	"context"
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

	tasks              schemas.Tasks
	tasksMutex         sync.RWMutex
	executedTasksCount uint64
}

// SetProject ..
func (l *Local) SetProject(_ context.Context, p schemas.Project) error {
	l.projectsMutex.Lock()
	defer l.projectsMutex.Unlock()

	l.projects[p.Key()] = p

	return nil
}

// DelProject ..
func (l *Local) DelProject(_ context.Context, k schemas.ProjectKey) error {
	l.projectsMutex.Lock()
	defer l.projectsMutex.Unlock()

	delete(l.projects, k)

	return nil
}

// GetProject ..
func (l *Local) GetProject(ctx context.Context, p *schemas.Project) error {
	exists, _ := l.ProjectExists(ctx, p.Key())

	if exists {
		l.projectsMutex.RLock()
		*p = l.projects[p.Key()]
		l.projectsMutex.RUnlock()
	}

	return nil
}

// ProjectExists ..
func (l *Local) ProjectExists(_ context.Context, k schemas.ProjectKey) (bool, error) {
	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	_, ok := l.projects[k]

	return ok, nil
}

// Projects ..
func (l *Local) Projects(_ context.Context) (projects schemas.Projects, err error) {
	projects = make(schemas.Projects)

	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	for k, v := range l.projects {
		projects[k] = v
	}

	return
}

// ProjectsCount ..
func (l *Local) ProjectsCount(_ context.Context) (int64, error) {
	l.projectsMutex.RLock()
	defer l.projectsMutex.RUnlock()

	return int64(len(l.projects)), nil
}

// SetEnvironment ..
func (l *Local) SetEnvironment(_ context.Context, environment schemas.Environment) error {
	l.environmentsMutex.Lock()
	defer l.environmentsMutex.Unlock()

	l.environments[environment.Key()] = environment

	return nil
}

// DelEnvironment ..
func (l *Local) DelEnvironment(_ context.Context, k schemas.EnvironmentKey) error {
	l.environmentsMutex.Lock()
	defer l.environmentsMutex.Unlock()

	delete(l.environments, k)

	return nil
}

// GetEnvironment ..
func (l *Local) GetEnvironment(ctx context.Context, environment *schemas.Environment) error {
	exists, _ := l.EnvironmentExists(ctx, environment.Key())

	if exists {
		l.environmentsMutex.RLock()
		*environment = l.environments[environment.Key()]
		l.environmentsMutex.RUnlock()
	}

	return nil
}

// EnvironmentExists ..
func (l *Local) EnvironmentExists(_ context.Context, k schemas.EnvironmentKey) (bool, error) {
	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	_, ok := l.environments[k]

	return ok, nil
}

// Environments ..
func (l *Local) Environments(_ context.Context) (environments schemas.Environments, err error) {
	environments = make(schemas.Environments)

	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	for k, v := range l.environments {
		environments[k] = v
	}

	return
}

// EnvironmentsCount ..
func (l *Local) EnvironmentsCount(_ context.Context) (int64, error) {
	l.environmentsMutex.RLock()
	defer l.environmentsMutex.RUnlock()

	return int64(len(l.environments)), nil
}

// SetRef ..
func (l *Local) SetRef(_ context.Context, ref schemas.Ref) error {
	l.refsMutex.Lock()
	defer l.refsMutex.Unlock()

	l.refs[ref.Key()] = ref

	return nil
}

// DelRef ..
func (l *Local) DelRef(_ context.Context, k schemas.RefKey) error {
	l.refsMutex.Lock()
	defer l.refsMutex.Unlock()

	delete(l.refs, k)

	return nil
}

// GetRef ..
func (l *Local) GetRef(ctx context.Context, ref *schemas.Ref) error {
	exists, _ := l.RefExists(ctx, ref.Key())

	if exists {
		l.refsMutex.RLock()
		*ref = l.refs[ref.Key()]
		l.refsMutex.RUnlock()
	}

	return nil
}

// RefExists ..
func (l *Local) RefExists(_ context.Context, k schemas.RefKey) (bool, error) {
	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	_, ok := l.refs[k]

	return ok, nil
}

// Refs ..
func (l *Local) Refs(_ context.Context) (refs schemas.Refs, err error) {
	refs = make(schemas.Refs)

	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	for k, v := range l.refs {
		refs[k] = v
	}

	return
}

// RefsCount ..
func (l *Local) RefsCount(_ context.Context) (int64, error) {
	l.refsMutex.RLock()
	defer l.refsMutex.RUnlock()

	return int64(len(l.refs)), nil
}

// SetMetric ..
func (l *Local) SetMetric(_ context.Context, m schemas.Metric) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	l.metrics[m.Key()] = m

	return nil
}

// DelMetric ..
func (l *Local) DelMetric(_ context.Context, k schemas.MetricKey) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	delete(l.metrics, k)

	return nil
}

// GetMetric ..
func (l *Local) GetMetric(ctx context.Context, m *schemas.Metric) error {
	exists, _ := l.MetricExists(ctx, m.Key())

	if exists {
		l.metricsMutex.RLock()
		*m = l.metrics[m.Key()]
		l.metricsMutex.RUnlock()
	}

	return nil
}

// MetricExists ..
func (l *Local) MetricExists(_ context.Context, k schemas.MetricKey) (bool, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	_, ok := l.metrics[k]

	return ok, nil
}

// Metrics ..
func (l *Local) Metrics(_ context.Context) (metrics schemas.Metrics, err error) {
	metrics = make(schemas.Metrics)

	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	for k, v := range l.metrics {
		metrics[k] = v
	}

	return
}

// MetricsCount ..
func (l *Local) MetricsCount(_ context.Context) (int64, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	return int64(len(l.metrics)), nil
}

// isTaskAlreadyQueued assess if a task is already queued or not.
func (l *Local) isTaskAlreadyQueued(tt schemas.TaskType, uniqueID string) bool {
	l.tasksMutex.Lock()
	defer l.tasksMutex.Unlock()

	if l.tasks == nil {
		l.tasks = make(map[schemas.TaskType]map[string]interface{})
	}

	taskTypeQueue, ok := l.tasks[tt]
	if !ok {
		l.tasks[tt] = make(map[string]interface{})

		return false
	}

	if _, alreadyQueued := taskTypeQueue[uniqueID]; alreadyQueued {
		return true
	}

	return false
}

// QueueTask registers that we are queueing the task.
// It returns true if it managed to schedule it, false if it was already scheduled.
func (l *Local) QueueTask(_ context.Context, tt schemas.TaskType, uniqueID, _ string) (bool, error) {
	if !l.isTaskAlreadyQueued(tt, uniqueID) {
		l.tasksMutex.Lock()
		defer l.tasksMutex.Unlock()

		l.tasks[tt][uniqueID] = nil

		return true, nil
	}

	return false, nil
}

// UnqueueTask removes the task from the tracker.
func (l *Local) UnqueueTask(_ context.Context, tt schemas.TaskType, uniqueID string) error {
	if l.isTaskAlreadyQueued(tt, uniqueID) {
		l.tasksMutex.Lock()
		defer l.tasksMutex.Unlock()

		delete(l.tasks[tt], uniqueID)

		l.executedTasksCount++
	}

	return nil
}

// CurrentlyQueuedTasksCount ..
func (l *Local) CurrentlyQueuedTasksCount(_ context.Context) (count uint64, err error) {
	l.tasksMutex.RLock()
	defer l.tasksMutex.RUnlock()

	for _, t := range l.tasks {
		count += uint64(len(t))
	}

	return
}

// ExecutedTasksCount ..
func (l *Local) ExecutedTasksCount(_ context.Context) (uint64, error) {
	l.tasksMutex.RLock()
	defer l.tasksMutex.RUnlock()

	return l.executedTasksCount, nil
}
