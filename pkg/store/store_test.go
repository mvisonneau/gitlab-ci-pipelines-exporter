package store

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

var testCtx = context.Background()

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
	}

	assert.Equal(t, expectedValue, NewRedisStore(redisClient))
}

func TestNew(t *testing.T) {
	localStore := New(testCtx, nil, config.Projects{})
	assert.IsType(t, &Local{}, localStore)

	redisClient := redis.NewClient(&redis.Options{})
	redisStore := New(testCtx, redisClient, config.Projects{})
	assert.IsType(t, &Redis{}, redisStore)

	localStore = New(testCtx, nil, config.Projects{
		{
			Name: "foo",
		},
		{
			Name: "foo",
		},
		{
			Name: "bar",
		},
	})
	count, _ := localStore.ProjectsCount(testCtx)
	assert.Equal(t, int64(2), count)
}
