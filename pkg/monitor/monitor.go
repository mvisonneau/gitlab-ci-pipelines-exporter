package monitor

import "time"

type TaskSchedulingStatus struct {
	Last time.Time
	Next time.Time
}
