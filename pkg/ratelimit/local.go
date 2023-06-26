package ratelimit

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Local ..
type Local struct {
	*rate.Limiter
}

// NewLocalLimiter ..
func NewLocalLimiter(maximumRPS, burstableRPS int) Limiter {
	return Local{
		rate.NewLimiter(rate.Limit(maximumRPS), burstableRPS),
	}
}

// Take ..
func (l Local) Take(ctx context.Context) time.Duration {
	start := time.Now()

	if err := l.Limiter.Wait(ctx); err != nil {
		log.WithContext(ctx).
			WithError(err).
			Fatal()
	}

	return start.Sub(time.Now())
}
