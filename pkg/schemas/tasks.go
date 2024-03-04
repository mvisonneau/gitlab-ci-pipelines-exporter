package schemas

// TaskType represents the type of a task.
type TaskType string

const (
	// TaskTypePullProject ..
	TaskTypePullProject TaskType = "PullProject"

	// TaskTypePullProjectsFromWildcard ..
	TaskTypePullProjectsFromWildcard TaskType = "PullProjectsFromWildcard"

	// TaskTypePullProjectsFromWildcards ..
	TaskTypePullProjectsFromWildcards TaskType = "PullProjectsFromWildcards"

	// TaskTypePullEnvironmentsFromProject ..
	TaskTypePullEnvironmentsFromProject TaskType = "PullEnvironmentsFromProject"

	// TaskTypePullEnvironmentsFromProjects ..
	TaskTypePullEnvironmentsFromProjects TaskType = "PullEnvironmentsFromProjects"

	// TaskTypePullEnvironmentMetrics ..
	TaskTypePullEnvironmentMetrics TaskType = "PullEnvironmentMetrics"

	// TaskTypePullMetrics ..
	TaskTypePullMetrics TaskType = "PullMetrics"

	// TaskTypePullRefsFromProject ..
	TaskTypePullRefsFromProject TaskType = "PullRefsFromProject"

	// TaskTypePullRefsFromProjects ..
	TaskTypePullRefsFromProjects TaskType = "PullRefsFromProjects"

	// TaskTypePullRefMetrics ..
	TaskTypePullRefMetrics TaskType = "PullRefMetrics"

	// TaskTypeGarbageCollectProjects ..
	TaskTypeGarbageCollectProjects TaskType = "GarbageCollectProjects"

	// TaskTypeGarbageCollectEnvironments ..
	TaskTypeGarbageCollectEnvironments TaskType = "GarbageCollectEnvironments"

	// TaskTypeGarbageCollectRefs ..
	TaskTypeGarbageCollectRefs TaskType = "GarbageCollectRefs"

	// TaskTypeGarbageCollectMetrics ..
	TaskTypeGarbageCollectMetrics TaskType = "GarbageCollectMetrics"
)

// Tasks can be used to keep track of tasks.
type Tasks map[TaskType]map[string]interface{}
