package monitor

import "time"

// Status ..
type Status struct {
	GitLabAPIUsage          float64 // ok
	GitLabAPIRequestsCount  uint64
	GitLabAPIRateLimit      float64 // ok
	GitLabAPILimitRemaining int
	TasksBufferUsage        float64 // ok
	TasksExecutedCount      uint64  // ok
	Projects                EntityStatus
	Refs                    EntityStatus
	Envs                    EntityStatus
	Metrics                 EntityStatus
}

// EntityStatus ..
type EntityStatus struct {
	Count    int64 // ok
	LastGC   time.Time
	LastPull time.Time
	NextGC   time.Time
	NextPull time.Time
}

// TaskSchedulingStatus ..
type TaskSchedulingStatus struct {
	Last time.Time
	Next time.Time
}
