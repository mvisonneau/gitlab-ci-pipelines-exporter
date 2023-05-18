package ratelimit

import (
	"context"
	"time"

	localRatelimit "go.uber.org/ratelimit"
)

// Local ..
type Local struct {
	localRatelimit.Limiter
}

// NewLocalLimiter ..
func NewLocalLimiter(maxRPS int, TimeWindow int) Limiter {
	return Local{
		localRatelimit.New(maxRPS, localRatelimit.Per(time.Duration(TimeWindow)*time.Second)),
	}
}

// Take ..
func (l Local) Take(_ context.Context) time.Time {
	return l.Limiter.Take()
}
