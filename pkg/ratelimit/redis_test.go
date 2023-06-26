package ratelimit

import (
	"testing"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLimiter(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	l := NewRedisLimiter(
		redisClient,
		10,
	)

	expectedValue := Redis{
		Limiter: redis_rate.NewLimiter(redisClient),
		MaxRPS:  10,
	}

	assert.Equal(t, expectedValue, l)
}
