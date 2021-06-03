package store

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalStore(t *testing.T) {
	expectedValue := &Local{
		projects:     make(schemas.Projects),
		environments: make(schemas.Environments),
		refs:         make(schemas.Refs),
		metrics:      make(schemas.Metrics),
	}
	assert.Equal(t, expectedValue, NewLocalStore())
}

func TestNewRedisStore(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	expectedValue := &Redis{
		Client: redisClient,
		ctx:    context.TODO(),
	}

	assert.Equal(t, expectedValue, NewRedisStore(redisClient))
}
