package exporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func getMockedGitlabClient() (*http.ServeMux, *httptest.Server, *gitlab.Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	opts := []goGitlab.ClientOptionFunc{
		goGitlab.WithBaseURL(server.URL),
		goGitlab.WithoutRetries(),
	}

	gc, _ := goGitlab.NewClient("", opts...)

	c := &gitlab.Client{
		Client:      gc,
		RateLimiter: ratelimit.NewLocalLimiter(100),
	}

	return mux, server, c
}

func TestConfigureGitlabClient(t *testing.T) {
	gc, err := goGitlab.NewClient("")
	assert.NoError(t, err)

	c := &gitlab.Client{
		Client:      gc,
		RateLimiter: ratelimit.NewLocalLimiter(10),
	}
	ConfigureGitlabClient(c)
	assert.Equal(t, gitlabClient, c)
}

func TestConfigureRedisClient(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	assert.NoError(t, ConfigureRedisClient(c))
	assert.Equal(t, redisClient, c)

	s.Close()
	assert.Error(t, ConfigureRedisClient(c))
}

func TestConfigurePollingQueue(t *testing.T) {
	// TODO: Test with redis client, miniredis does not seem to support it yet
	redisClient = nil
	ConfigurePollingQueue()
	assert.Equal(t, "poll", pollingQueue.Options().Name)
}

func TestConfigureStore(t *testing.T) {
	Config = schemas.Config{
		Projects: []schemas.Project{
			{
				Name: "foo/bar",
			},
		},
	}

	// Test with local storage
	redisClient = nil
	ConfigureStore()
	assert.NotNil(t, store)

	projects, err := store.Projects()
	assert.NoError(t, err)

	expectedProjects := schemas.Projects{
		"F83q76XMYCJIHIJOFaR6dyb1k90=": schemas.Project{
			Name: "foo/bar",
		},
	}
	assert.Equal(t, expectedProjects, projects)

	// Test with redis storage
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	assert.NoError(t, ConfigureRedisClient(c))

	ConfigureStore()
	projects, err = store.Projects()
	assert.NoError(t, err)
	assert.Equal(t, expectedProjects, projects)
}

func TestProcessPollingQueue(t *testing.T) {
	// TODO: Test with redis client, miniredis does not seem to support it yet
	redisClient = nil
	ProcessPollingQueue(context.TODO())
}
