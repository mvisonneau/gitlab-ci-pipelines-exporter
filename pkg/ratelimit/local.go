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
func NewLocalLimiter(maxRPS int) Limiter {
	return Local{
		localRatelimit.New(maxRPS),
	}
}

// Take ..
func (l Local) Take(_ context.Context) time.Time {
	return l.Limiter.Take()
}
