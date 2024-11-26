package ratelimit

import (
	"context"
	"time"
)

// Limiter ..
type Limiter interface {
	Take(ctx context.Context) time.Duration
}

// Take ..
func Take(ctx context.Context, l Limiter) {
	l.Take(ctx)
}
