package ratelimit

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func MeasureTakeDuration(l Limiter) int64 {
	start := time.Now()

	Take(context.TODO(), l)

	return int64(time.Since(start))
}

func TestLocalTake(t *testing.T) {
	l := NewLocalLimiter(1, 1)

	assert.LessOrEqual(t, MeasureTakeDuration(l), int64(100*time.Millisecond))
	assert.GreaterOrEqual(t, MeasureTakeDuration(l), int64(time.Second))
}

func TestRedisTake(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	defer s.Close()

	l := NewRedisLimiter(
		redis.NewClient(&redis.Options{Addr: s.Addr()}),
		1,
	)

	assert.LessOrEqual(t, MeasureTakeDuration(l), int64(250*time.Millisecond))
	assert.GreaterOrEqual(t, MeasureTakeDuration(l), int64(900*time.Millisecond))
}

func TestRedisTakeError(t *testing.T) {
	if os.Getenv("SHOULD_ERROR") == "1" {
		l := NewRedisLimiter(
			redis.NewClient(&redis.Options{Addr: "doesnotexist"}),
			1,
		)

		Take(context.TODO(), l)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRedisTakeError")
	cmd.Env = append(os.Environ(), "SHOULD_ERROR=1")

	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatal("process ran successfully, wanted exit status 1")
}
