package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	redisProjectsKey     string = `projects`
	redisProjectsRefsKey string = `projectsRefs`
	redisMetricsKey      string = `metrics`
)

// Redis ..
type Redis struct {
	*redis.Client

	ctx context.Context
}

// SetProject ..
func (r *Redis) SetProject(p schemas.Project) error {
	marshalledProject, err := msgpack.Marshal(p)
	if err != nil {
		return err
	}

	if _, err := r.HSet(r.ctx, redisProjectsKey, string(p.Key()), marshalledProject).Result(); err != nil {
		return err
	}
	return nil
}

// DelProject ..
func (r *Redis) DelProject(k schemas.ProjectKey) error {
	if _, err := r.HDel(r.ctx, redisProjectsKey, string(k)).Result(); err != nil {
		return err
	}
	return nil
}

// ProjectExists ..
func (r *Redis) ProjectExists(k schemas.ProjectKey) (bool, error) {
	pExists, err := r.HExists(r.ctx, redisProjectsKey, string(k)).Result()
	if err != nil {
		return false, err
	}
	return pExists, nil
}

// Projects ..
func (r *Redis) Projects() (schemas.Projects, error) {
	projects := schemas.Projects{}
	marshalledProjects, err := r.HGetAll(r.ctx, redisProjectsKey).Result()
	if err != nil {
		return projects, err
	}

	for stringProjectKey, marshalledProject := range marshalledProjects {
		p := schemas.Project{}

		if err := msgpack.Unmarshal([]byte(marshalledProject), &p); err != nil {
			return projects, err
		}
		projects[schemas.ProjectKey(stringProjectKey)] = p
	}

	return projects, nil
}

// ProjectsCount ..
func (r *Redis) ProjectsCount() (int64, error) {
	return r.HLen(r.ctx, redisProjectsKey).Result()
}

// SetProjectRef ..
func (r *Redis) SetProjectRef(pr schemas.ProjectRef) error {
	marshalledProjectRef, err := msgpack.Marshal(pr)
	if err != nil {
		return err
	}

	if _, err := r.HSet(r.ctx, redisProjectsRefsKey, string(pr.Key()), marshalledProjectRef).Result(); err != nil {
		return err
	}
	return nil
}

// DelProjectRef ..
func (r *Redis) DelProjectRef(k schemas.ProjectRefKey) error {
	if _, err := r.HDel(r.ctx, redisProjectsRefsKey, string(k)).Result(); err != nil {
		return err
	}
	return nil
}

// ProjectRefExists ..
func (r *Redis) ProjectRefExists(k schemas.ProjectRefKey) (bool, error) {
	pExists, err := r.HExists(r.ctx, redisProjectsRefsKey, string(k)).Result()
	if err != nil {
		return false, err
	}
	return pExists, nil
}

// ProjectsRefs ..
func (r *Redis) ProjectsRefs() (schemas.ProjectsRefs, error) {
	projectsRefs := schemas.ProjectsRefs{}
	marshalledProjects, err := r.HGetAll(r.ctx, redisProjectsRefsKey).Result()
	if err != nil {
		return projectsRefs, err
	}

	for stringProjectRefKey, marshalledProjectRef := range marshalledProjects {
		p := schemas.ProjectRef{}

		if err := msgpack.Unmarshal([]byte(marshalledProjectRef), &p); err != nil {
			return projectsRefs, err
		}
		projectsRefs[schemas.ProjectRefKey(stringProjectRefKey)] = p
	}

	return projectsRefs, nil
}

// ProjectsRefsCount ..
func (r *Redis) ProjectsRefsCount() (int64, error) {
	return r.HLen(r.ctx, redisProjectsRefsKey).Result()
}

// SetMetric ..
func (r *Redis) SetMetric(m schemas.Metric) error {
	marshalledMetric, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	if _, err := r.HSet(r.ctx, redisMetricsKey, string(m.Key()), marshalledMetric).Result(); err != nil {
		return err
	}
	return nil
}

// DelMetric ..
func (r *Redis) DelMetric(k schemas.MetricKey) error {
	if _, err := r.HDel(r.ctx, redisMetricsKey, string(k)).Result(); err != nil {
		return err
	}
	return nil
}

// MetricExists ..
func (r *Redis) MetricExists(k schemas.MetricKey) (bool, error) {
	exists, err := r.HExists(r.ctx, redisProjectsRefsKey, string(k)).Result()
	if err != nil {
		return false, err
	}
	return exists, nil
}

// PullMetricValue ..
func (r *Redis) PullMetricValue(m *schemas.Metric) error {
	exists, err := r.MetricExists(m.Key())
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	k := m.Key()
	marshalledMetric, err := r.HGet(r.ctx, redisMetricsKey, string(k)).Result()
	if err != nil {
		return err
	}

	storedMetric := schemas.Metric{}
	if err = msgpack.Unmarshal([]byte(marshalledMetric), &storedMetric); err != nil {
		return err
	}

	m.Value = storedMetric.Value
	return nil
}

// Metrics ..
func (r *Redis) Metrics() (schemas.Metrics, error) {
	metrics := schemas.Metrics{}
	marshalledMetrics, err := r.HGetAll(r.ctx, redisMetricsKey).Result()
	if err != nil {
		return metrics, err
	}

	for stringMetricKey, marshalledMetric := range marshalledMetrics {
		m := schemas.Metric{}

		if err := msgpack.Unmarshal([]byte(marshalledMetric), &m); err != nil {
			return metrics, err
		}
		metrics[schemas.MetricKey(stringMetricKey)] = m
	}

	return metrics, nil
}

// MetricsCount ..
func (r *Redis) MetricsCount() (int64, error) {
	return r.HLen(r.ctx, redisMetricsKey).Result()
}
