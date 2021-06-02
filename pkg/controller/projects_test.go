package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPullProjectsFromWildcard(t *testing.T) {
	c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"path_with_namespace":"foo","jobs_enabled":false},{"id":2,"path_with_namespace":"bar","jobs_enabled":true}]`)
		})

	w := config.Wildcard{}
	assert.NoError(t, c.PullProjectsFromWildcard(context.Background(), w))

	projects, _ := c.Store.Projects()
	expectedProjects := config.Projects{
		"1996459178": config.Project{
			Name: "bar",
		},
	}
	assert.Equal(t, expectedProjects, projects)
}
