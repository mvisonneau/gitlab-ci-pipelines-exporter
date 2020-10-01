package ratelimit

import "time"

// Limiter ..
type Limiter interface {
	Take() time.Time
}
