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

	topicsIdentifier := 0

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `[{"id":2,"path_with_namespace":"bar","jobs_enabled":true,"topics":["foo%d","bar%d"]}]`, topicsIdentifier, topicsIdentifier)
			topicsIdentifier += 1
		})

	w := config.NewWildcard()
	assert.NoError(t, c.PullProjectsFromWildcard(ctx, w))

	projects, _ := c.Store.Projects(ctx)
	p1 := schemas.NewProject("bar", []string{"foo0", "bar0"})
	p2 := schemas.NewProject("bar", []string{"foo1", "bar1"})

	expectedProjects := schemas.Projects{
		p1.Key(): p1,
	}
	assert.Equal(t, expectedProjects, projects)

	expectedUpdatedProjects := schemas.Projects{
		p2.Key(): p2,
	}

	// Pull projects again, which will have topics updated
	assert.NoError(t, c.PullProjectsFromWildcard(ctx, w))
	projects, _ = c.Store.Projects(ctx)
	assert.Equal(t, expectedUpdatedProjects, projects)
}

func TestPullProjects(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	topicsIdentifier := 0

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%s", "foo%2Fbar"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"id":2,"path_with_namespace":"bar","jobs_enabled":true,"topics":["foo%d","bar%d"]}`, topicsIdentifier, topicsIdentifier)
			topicsIdentifier += 1
		})

	w := config.NewProject("foo/bar")
	assert.NoError(t, c.PullProject(ctx, w))

	projects, _ := c.Store.Projects(ctx)
	p1 := schemas.NewProject("bar", []string{"foo0", "bar0"})
	p2 := schemas.NewProject("bar", []string{"foo1", "bar1"})

	expectedProjects := schemas.Projects{
		p1.Key(): p1,
	}
	assert.Equal(t, expectedProjects, projects)

	expectedUpdatedProjects := schemas.Projects{
		p2.Key(): p2,
	}

	// Pull projects again, which will have topics updated
	assert.NoError(t, c.PullProject(ctx, w))
	projects, _ = c.Store.Projects(ctx)
	assert.Equal(t, expectedUpdatedProjects, projects)
}
