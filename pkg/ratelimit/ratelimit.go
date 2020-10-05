package ratelimit

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// Limiter ..
type Limiter interface {
	Take() time.Time
}

// Take ..
func Take(l Limiter) {
	now := time.Now()
	throttled := l.Take()
	if throttled.Sub(now).Milliseconds() > 10 {
		log.WithFields(
			log.Fields{
				"for": throttled.Sub(now),
			},
		).Debug("throttled GitLab requests")
	}
	return
}
