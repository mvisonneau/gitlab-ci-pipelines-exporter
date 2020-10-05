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
	Context context.Context
	MaxRPS  int
}

// NewRedisLimiter ..
func NewRedisLimiter(ctx context.Context, redisClient *redis.Client, maxRPS int) Limiter {
	return Redis{
		Limiter: redis_rate.NewLimiter(redisClient),
		Context: ctx,
		MaxRPS:  maxRPS,
	}
}

// Take ..
func (r Redis) Take() time.Time {
	res, err := r.Allow(r.Context, redisKey, redis_rate.PerSecond(r.MaxRPS))
	if err != nil {
		log.Fatalf(err.Error())
	}
	time.Sleep(res.RetryAfter)
	return time.Now()
}
