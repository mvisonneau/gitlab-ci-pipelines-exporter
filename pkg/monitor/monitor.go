package monitor

import "time"

// TaskSchedulingStatus reports the status of a scheduled job.
type TaskSchedulingStatus struct {
	Last time.Time
	Next time.Time
}
