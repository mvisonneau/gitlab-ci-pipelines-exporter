package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

const (
	redisProjectsKey           string = `projects`
	redisEnvironmentsKey       string = `environments`
	redisRefsKey               string = `refs`
	redisMetricsKey            string = `metrics`
	redisTaskKey               string = `task`
	redisTasksExecutedCountKey string = `tasksExecutedCount`
	redisKeepaliveKey          string = `keepalive`
)

// Redis ..
type Redis struct {
	*redis.Client
}

// SetProject ..
func (r *Redis) SetProject(ctx context.Context, p schemas.Project) error {
	marshalledProject, err := msgpack.Marshal(p)
	if err != nil {
		return err
	}

	_, err = r.HSet(ctx, redisProjectsKey, string(p.Key()), marshalledProject).Result()

	return err
}

// DelProject ..
func (r *Redis) DelProject(ctx context.Context, k schemas.ProjectKey) error {
	_, err := r.HDel(ctx, redisProjectsKey, string(k)).Result()

	return err
}

// GetProject ..
func (r *Redis) GetProject(ctx context.Context, p *schemas.Project) error {
	exists, err := r.ProjectExists(ctx, p.Key())
	if err != nil {
		return err
	}

	if exists {
		k := p.Key()

		marshalledProject, err := r.HGet(ctx, redisProjectsKey, string(k)).Result()
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
func (r *Redis) ProjectExists(ctx context.Context, k schemas.ProjectKey) (bool, error) {
	return r.HExists(ctx, redisProjectsKey, string(k)).Result()
}

// Projects ..
func (r *Redis) Projects(ctx context.Context) (schemas.Projects, error) {
	projects := schemas.Projects{}

	marshalledProjects, err := r.HGetAll(ctx, redisProjectsKey).Result()
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
func (r *Redis) ProjectsCount(ctx context.Context) (int64, error) {
	return r.HLen(ctx, redisProjectsKey).Result()
}

// SetEnvironment ..
func (r *Redis) SetEnvironment(ctx context.Context, e schemas.Environment) error {
	marshalledEnvironment, err := msgpack.Marshal(e)
	if err != nil {
		return err
	}

	_, err = r.HSet(ctx, redisEnvironmentsKey, string(e.Key()), marshalledEnvironment).Result()

	return err
}

// DelEnvironment ..
func (r *Redis) DelEnvironment(ctx context.Context, k schemas.EnvironmentKey) error {
	_, err := r.HDel(ctx, redisEnvironmentsKey, string(k)).Result()

	return err
}

// GetEnvironment ..
func (r *Redis) GetEnvironment(ctx context.Context, e *schemas.Environment) error {
	exists, err := r.EnvironmentExists(ctx, e.Key())
	if err != nil {
		return err
	}

	if exists {
		k := e.Key()

		marshalledEnvironment, err := r.HGet(ctx, redisEnvironmentsKey, string(k)).Result()
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
func (r *Redis) EnvironmentExists(ctx context.Context, k schemas.EnvironmentKey) (bool, error) {
	return r.HExists(ctx, redisEnvironmentsKey, string(k)).Result()
}

// Environments ..
func (r *Redis) Environments(ctx context.Context) (schemas.Environments, error) {
	environments := schemas.Environments{}

	marshalledProjects, err := r.HGetAll(ctx, redisEnvironmentsKey).Result()
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
func (r *Redis) EnvironmentsCount(ctx context.Context) (int64, error) {
	return r.HLen(ctx, redisEnvironmentsKey).Result()
}

// SetRef ..
func (r *Redis) SetRef(ctx context.Context, ref schemas.Ref) error {
	marshalledRef, err := msgpack.Marshal(ref)
	if err != nil {
		return err
	}

	_, err = r.HSet(ctx, redisRefsKey, string(ref.Key()), marshalledRef).Result()

	return err
}

// DelRef ..
func (r *Redis) DelRef(ctx context.Context, k schemas.RefKey) error {
	_, err := r.HDel(ctx, redisRefsKey, string(k)).Result()

	return err
}

// GetRef ..
func (r *Redis) GetRef(ctx context.Context, ref *schemas.Ref) error {
	exists, err := r.RefExists(ctx, ref.Key())
	if err != nil {
		return err
	}

	if exists {
		k := ref.Key()

		marshalledRef, err := r.HGet(ctx, redisRefsKey, string(k)).Result()
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
func (r *Redis) RefExists(ctx context.Context, k schemas.RefKey) (bool, error) {
	return r.HExists(ctx, redisRefsKey, string(k)).Result()
}

// Refs ..
func (r *Redis) Refs(ctx context.Context) (schemas.Refs, error) {
	refs := schemas.Refs{}

	marshalledProjects, err := r.HGetAll(ctx, redisRefsKey).Result()
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
func (r *Redis) RefsCount(ctx context.Context) (int64, error) {
	return r.HLen(ctx, redisRefsKey).Result()
}

// SetMetric ..
func (r *Redis) SetMetric(ctx context.Context, m schemas.Metric) error {
	marshalledMetric, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	_, err = r.HSet(ctx, redisMetricsKey, string(m.Key()), marshalledMetric).Result()

	return err
}

// DelMetric ..
func (r *Redis) DelMetric(ctx context.Context, k schemas.MetricKey) error {
	_, err := r.HDel(ctx, redisMetricsKey, string(k)).Result()

	return err
}

// MetricExists ..
func (r *Redis) MetricExists(ctx context.Context, k schemas.MetricKey) (bool, error) {
	return r.HExists(ctx, redisMetricsKey, string(k)).Result()
}

// GetMetric ..
func (r *Redis) GetMetric(ctx context.Context, m *schemas.Metric) error {
	exists, err := r.MetricExists(ctx, m.Key())
	if err != nil {
		return err
	}

	if exists {
		k := m.Key()

		marshalledMetric, err := r.HGet(ctx, redisMetricsKey, string(k)).Result()
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
func (r *Redis) Metrics(ctx context.Context) (schemas.Metrics, error) {
	metrics := schemas.Metrics{}

	marshalledMetrics, err := r.HGetAll(ctx, redisMetricsKey).Result()
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
func (r *Redis) MetricsCount(ctx context.Context) (int64, error) {
	return r.HLen(ctx, redisMetricsKey).Result()
}

// SetKeepalive sets a key with an UUID corresponding to the currently running process.
func (r *Redis) SetKeepalive(ctx context.Context, uuid string, ttl time.Duration) (bool, error) {
	return r.SetNX(ctx, fmt.Sprintf("%s:%s", redisKeepaliveKey, uuid), nil, ttl).Result()
}

// KeepaliveExists returns whether a keepalive exists or not for a particular UUID.
func (r *Redis) KeepaliveExists(ctx context.Context, uuid string) (bool, error) {
	exists, err := r.Exists(ctx, fmt.Sprintf("%s:%s", redisKeepaliveKey, uuid)).Result()

	return exists == 1, err
}

func getRedisQueueKey(tt schemas.TaskType, taskUUID string) string {
	return fmt.Sprintf("%s:%v:%s", redisTaskKey, tt, taskUUID)
}

// QueueTask registers that we are queueing the task.
// It returns true if it managed to schedule it, false if it was already scheduled.
func (r *Redis) QueueTask(ctx context.Context, tt schemas.TaskType, taskUUID, processUUID string) (set bool, err error) {
	k := getRedisQueueKey(tt, taskUUID)

	// We attempt to set the key, if it already exists, we do not overwrite it
	set, err = r.SetNX(ctx, k, processUUID, 0).Result()
	if err != nil || set {
		return
	}

	// If the key already exists, we want to check a couple of things
	// First, that the associated process UUID is the same as our current one
	var tpuuid string

	if tpuuid, err = r.Get(ctx, k).Result(); err != nil {
		return
	}

	// If it is not the case, we assess that the one being associated with the task lock
	// is still alive, otherwise we override the key and schedule the task
	if tpuuid != processUUID {
		var uuidIsAlive bool

		if uuidIsAlive, err = r.KeepaliveExists(ctx, tpuuid); err != nil {
			return
		}

		if !uuidIsAlive {
			if _, err = r.Set(ctx, k, processUUID, 0).Result(); err != nil {
				return
			}

			return true, nil
		}
	}

	return
}

// UnqueueTask removes the task from the tracker.
func (r *Redis) UnqueueTask(ctx context.Context, tt schemas.TaskType, taskUUID string) (err error) {
	var matched int64

	matched, err = r.Del(ctx, getRedisQueueKey(tt, taskUUID)).Result()
	if err != nil {
		return
	}

	if matched > 0 {
		_, err = r.Incr(ctx, redisTasksExecutedCountKey).Result()
	}

	return
}

// CurrentlyQueuedTasksCount ..
func (r *Redis) CurrentlyQueuedTasksCount(ctx context.Context) (count uint64, err error) {
	iter := r.Scan(ctx, 0, fmt.Sprintf("%s:*", redisTaskKey), 0).Iterator()
	for iter.Next(ctx) {
		count++
	}

	err = iter.Err()

	return
}

// ExecutedTasksCount ..
func (r *Redis) ExecutedTasksCount(ctx context.Context) (uint64, error) {
	countString, err := r.Get(ctx, redisTasksExecutedCountKey).Result()
	if err != nil {
		return 0, err
	}

	c, err := strconv.Atoi(countString)

	return uint64(c), err
}
