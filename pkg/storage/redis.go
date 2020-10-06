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

	_, err = r.HSet(r.ctx, redisProjectsKey, string(p.Key()), marshalledProject).Result()
	return err
}

// DelProject ..
func (r *Redis) DelProject(k schemas.ProjectKey) error {
	_, err := r.HDel(r.ctx, redisProjectsKey, string(k)).Result()
	return err
}

// GetProject ..
func (r *Redis) GetProject(p *schemas.Project) error {
	exists, err := r.ProjectExists(p.Key())
	if err != nil {
		return err
	}

	if exists {
		k := p.Key()
		marshalledProject, err := r.HGet(r.ctx, redisProjectsKey, string(k)).Result()
		if err != nil {
			return err
		}

		storedProject := schemas.Project{}
		if err = msgpack.Unmarshal([]byte(marshalledProject), &storedProject); err != nil {
			return err
		}

		*p = storedProject
	}

	return nil
}

// ProjectExists ..
func (r *Redis) ProjectExists(k schemas.ProjectKey) (bool, error) {
	return r.HExists(r.ctx, redisProjectsKey, string(k)).Result()
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

		if err = msgpack.Unmarshal([]byte(marshalledProject), &p); err != nil {
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

	_, err = r.HSet(r.ctx, redisProjectsRefsKey, string(pr.Key()), marshalledProjectRef).Result()
	return err
}

// DelProjectRef ..
func (r *Redis) DelProjectRef(k schemas.ProjectRefKey) error {
	_, err := r.HDel(r.ctx, redisProjectsRefsKey, string(k)).Result()
	return err
}

// GetProjectRef ..
func (r *Redis) GetProjectRef(pr *schemas.ProjectRef) error {
	exists, err := r.ProjectRefExists(pr.Key())
	if err != nil {
		return err
	}

	if exists {
		k := pr.Key()
		marshalledProjectRef, err := r.HGet(r.ctx, redisProjectsRefsKey, string(k)).Result()
		if err != nil {
			return err
		}

		storedProjectRef := schemas.ProjectRef{}
		if err = msgpack.Unmarshal([]byte(marshalledProjectRef), &storedProjectRef); err != nil {
			return err
		}

		*pr = storedProjectRef
	}

	return nil
}

// ProjectRefExists ..
func (r *Redis) ProjectRefExists(k schemas.ProjectRefKey) (bool, error) {
	return r.HExists(r.ctx, redisProjectsRefsKey, string(k)).Result()
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

		if err = msgpack.Unmarshal([]byte(marshalledProjectRef), &p); err != nil {
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

	_, err = r.HSet(r.ctx, redisMetricsKey, string(m.Key()), marshalledMetric).Result()
	return err
}

// DelMetric ..
func (r *Redis) DelMetric(k schemas.MetricKey) error {
	_, err := r.HDel(r.ctx, redisMetricsKey, string(k)).Result()
	return err
}

// MetricExists ..
func (r *Redis) MetricExists(k schemas.MetricKey) (bool, error) {
	return r.HExists(r.ctx, redisMetricsKey, string(k)).Result()
}

// GetMetric ..
func (r *Redis) GetMetric(m *schemas.Metric) error {
	exists, err := r.MetricExists(m.Key())
	if err != nil {
		return err
	}

	if exists {
		k := m.Key()
		marshalledMetric, err := r.HGet(r.ctx, redisMetricsKey, string(k)).Result()
		if err != nil {
			return err
		}

		storedMetric := schemas.Metric{}
		if err = msgpack.Unmarshal([]byte(marshalledMetric), &storedMetric); err != nil {
			return err
		}

		*m = storedMetric
	}

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
