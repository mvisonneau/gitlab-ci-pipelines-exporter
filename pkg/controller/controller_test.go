package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
)

func newMockedGitlabAPIServer() (mux *http.ServeMux, srv *httptest.Server) {
	mux = http.NewServeMux()
	srv = httptest.NewServer(mux)

	return
}

func newTestController(cfg config.Config) (ctx context.Context, c Controller, mux *http.ServeMux, srv *httptest.Server) {
	ctx = context.Background()
	mux, srv = newMockedGitlabAPIServer()

	cfg.Gitlab.URL = srv.URL
	if cfg.Gitlab.MaximumRequestsPerSecond < 1 {
		cfg.Gitlab.MaximumRequestsPerSecond = 1000
	}

	if cfg.Gitlab.BurstableRequestsPerSecond < 1 {
		cfg.Gitlab.BurstableRequestsPerSecond = 1
	}

	c, _ = New(context.Background(), cfg, "0.0.0-ci")

	return
}

func TestConfigureGitlab(t *testing.T) {
	c := Controller{}
	assert.NoError(t, c.configureGitlab(
		config.Gitlab{
			MaximumRequestsPerSecond: 5,
		},
		"0.0.0",
	))
	assert.NotNil(t, c.Gitlab)
}

// func TestConfigureRedisClient(t *testing.T) {

// 	s, err := miniredis.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer s.Close()

// 	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
// 	assert.NoError(t, ConfigureRedisClient(c))
// 	assert.Equal(t, redisClient, c)

// 	s.Close()
// 	assert.Error(t, ConfigureRedisClient(c))
// }

// func TestConfigureStore(t *testing.T) {
// 		cfg = config.Config{
// 		Projects: []config.Project{
// 			{
// 				Name: "foo/bar",
// 			},
// 		},
// 	}

// 	// Test with local storage
// 	configureStore()
// 	assert.NotNil(t, store)

// 	projects, err := store.Projects()
// 	assert.NoError(t, err)

// 	expectedProjects := config.Projects{
// 		"3861188962": config.Project{
// 			Name: "foo/bar",
// 		},
// 	}
// 	assert.Equal(t, expectedProjects, projects)

// 	// Test with redis storage
// 	s, err := miniredis.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer s.Close()

// 	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
// 	assert.NoError(t, ConfigureRedisClient(c))

// 	configureStore()
// 	projects, err = store.Projects()
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedProjects, projects)
// }
