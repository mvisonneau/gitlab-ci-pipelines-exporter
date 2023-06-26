package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestPullProjectsFromWildcard(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":2,"path_with_namespace":"bar","jobs_enabled":true}]`)
		})

	w := config.NewWildcard()
	assert.NoError(t, c.PullProjectsFromWildcard(ctx, w))

	projects, _ := c.Store.Projects(ctx)
	p1 := schemas.NewProject("bar")

	expectedProjects := schemas.Projects{
		p1.Key(): p1,
	}
	assert.Equal(t, expectedProjects, projects)
}
