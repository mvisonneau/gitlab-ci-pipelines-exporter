package ratelimit

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLimiter(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	l := NewRedisLimiter(
		context.Background(),
		redisClient,
		10,
	)

	expectedValue := Redis{
		Limiter: redis_rate.NewLimiter(redisClient),
		Context: context.Background(),
		MaxRPS:  10,
	}

	assert.Equal(t, expectedValue, l)
}
