package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	redisProjectsKey     string = `projects`
	redisEnvironmentsKey string = `environments`
	redisRefsKey         string = `refs`
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

		if err = msgpack.Unmarshal([]byte(marshalledProject), p); err != nil {
			return err
		}
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

// SetEnvironment ..
func (r *Redis) SetEnvironment(e schemas.Environment) error {
	marshalledEnvironment, err := msgpack.Marshal(e)
	if err != nil {
		return err
	}

	_, err = r.HSet(r.ctx, redisEnvironmentsKey, string(e.Key()), marshalledEnvironment).Result()
	return err
}

// DelEnvironment ..
func (r *Redis) DelEnvironment(k schemas.EnvironmentKey) error {
	_, err := r.HDel(r.ctx, redisEnvironmentsKey, string(k)).Result()
	return err
}

// GetEnvironment ..
func (r *Redis) GetEnvironment(e *schemas.Environment) error {
	exists, err := r.EnvironmentExists(e.Key())
	if err != nil {
		return err
	}

	if exists {
		k := e.Key()
		marshalledEnvironment, err := r.HGet(r.ctx, redisEnvironmentsKey, string(k)).Result()
		if err != nil {
			return err
		}

		if err = msgpack.Unmarshal([]byte(marshalledEnvironment), e); err != nil {
			return err
		}
	}

	return nil
}

// EnvironmentExists ..
func (r *Redis) EnvironmentExists(k schemas.EnvironmentKey) (bool, error) {
	return r.HExists(r.ctx, redisEnvironmentsKey, string(k)).Result()
}

// Environments ..
func (r *Redis) Environments() (schemas.Environments, error) {
	environments := schemas.Environments{}
	marshalledProjects, err := r.HGetAll(r.ctx, redisEnvironmentsKey).Result()
	if err != nil {
		return environments, err
	}

	for stringEnvironmentKey, marshalledEnvironment := range marshalledProjects {
		p := schemas.Environment{}

		if err = msgpack.Unmarshal([]byte(marshalledEnvironment), &p); err != nil {
			return environments, err
		}
		environments[schemas.EnvironmentKey(stringEnvironmentKey)] = p
	}

	return environments, nil
}

// EnvironmentsCount ..
func (r *Redis) EnvironmentsCount() (int64, error) {
	return r.HLen(r.ctx, redisEnvironmentsKey).Result()
}

// SetRef ..
func (r *Redis) SetRef(ref schemas.Ref) error {
	marshalledRef, err := msgpack.Marshal(ref)
	if err != nil {
		return err
	}

	_, err = r.HSet(r.ctx, redisRefsKey, string(ref.Key()), marshalledRef).Result()
	return err
}

// DelRef ..
func (r *Redis) DelRef(k schemas.RefKey) error {
	_, err := r.HDel(r.ctx, redisRefsKey, string(k)).Result()
	return err
}

// GetRef ..
func (r *Redis) GetRef(ref *schemas.Ref) error {
	exists, err := r.RefExists(ref.Key())
	if err != nil {
		return err
	}

	if exists {
		k := ref.Key()
		marshalledRef, err := r.HGet(r.ctx, redisRefsKey, string(k)).Result()
		if err != nil {
			return err
		}

		if err = msgpack.Unmarshal([]byte(marshalledRef), ref); err != nil {
			return err
		}
	}

	return nil
}

// RefExists ..
func (r *Redis) RefExists(k schemas.RefKey) (bool, error) {
	return r.HExists(r.ctx, redisRefsKey, string(k)).Result()
}

// Refs ..
func (r *Redis) Refs() (schemas.Refs, error) {
	refs := schemas.Refs{}
	marshalledProjects, err := r.HGetAll(r.ctx, redisRefsKey).Result()
	if err != nil {
		return refs, err
	}

	for stringRefKey, marshalledRef := range marshalledProjects {
		p := schemas.Ref{}

		if err = msgpack.Unmarshal([]byte(marshalledRef), &p); err != nil {
			return refs, err
		}
		refs[schemas.RefKey(stringRefKey)] = p
	}

	return refs, nil
}

// RefsCount ..
func (r *Redis) RefsCount() (int64, error) {
	return r.HLen(r.ctx, redisRefsKey).Result()
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

		if err = msgpack.Unmarshal([]byte(marshalledMetric), m); err != nil {
			return err
		}
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
