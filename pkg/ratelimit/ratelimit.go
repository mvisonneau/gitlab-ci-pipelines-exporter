package ratelimit

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

// Limiter ..
type Limiter interface {
	Take(context.Context) time.Time
}

// Take ..
func Take(ctx context.Context, l Limiter) {
	now := time.Now()
	throttled := l.Take(ctx)

	if throttled.Sub(now).Milliseconds() > 10 {
		log.WithFields(
			log.Fields{
				"for": throttled.Sub(now),
			},
		).Debug("throttled GitLab requests")
	}
}
