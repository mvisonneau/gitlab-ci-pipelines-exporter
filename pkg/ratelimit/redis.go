package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	log "github.com/sirupsen/logrus"
)

const redisKey string = `gcpe:gitlab:api`

// Redis ..
type Redis struct {
	*redis_rate.Limiter
	MaxRPS int
}

// NewRedisLimiter ..
func NewRedisLimiter(redisClient *redis.Client, maxRPS int) Limiter {
	return Redis{
		Limiter: redis_rate.NewLimiter(redisClient),
		MaxRPS:  maxRPS,
	}
}

// Take ..
func (r Redis) Take(ctx context.Context) time.Time {
	res, err := r.Allow(ctx, redisKey, redis_rate.PerSecond(r.MaxRPS))
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Fatal()
	}

	time.Sleep(res.RetryAfter)

	return time.Now()
}
