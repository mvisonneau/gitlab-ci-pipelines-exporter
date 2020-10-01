package ratelimit

import (
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
