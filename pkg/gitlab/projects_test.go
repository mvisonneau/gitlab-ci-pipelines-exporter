package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetProject(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	project := "foo/bar"
	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%s", project),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `{"id":1}`)
		})

	p, err := c.GetProject(project)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, 1, p.ID)
}

func TestListUserProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := schemas.Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string `yaml:"name"`
			Kind             string `yaml:"kind"`
			IncludeSubgroups bool   `yaml:"include_subgroups"`
		}{
			Name:             "foo",
			Kind:             "user",
			IncludeSubgroups: false,
		},
		Archived: false,
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.ListProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projects))
}

func TestListGroupProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := schemas.Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string `yaml:"name"`
			Kind             string `yaml:"kind"`
			IncludeSubgroups bool   `yaml:"include_subgroups"`
		}{
			Name:             "foo",
			Kind:             "group",
			IncludeSubgroups: false,
		},
		Archived: false,
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/groups/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.ListProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projects))
}

func TestListProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := schemas.Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string `yaml:"name"`
			Kind             string `yaml:"kind"`
			IncludeSubgroups bool   `yaml:"include_subgroups"`
		}{
			Name:             "",
			Kind:             "",
			IncludeSubgroups: false,
		},
		Archived: false,
	}

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.ListProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projects))
}

func TestListProjectsAPIError(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := schemas.Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string `yaml:"name"`
			Kind             string `yaml:"kind"`
			IncludeSubgroups bool   `yaml:"include_subgroups"`
		}{
			Name: "foo",
			Kind: "user",
		},
		Archived: false,
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		})

	_, err := c.ListProjects(w)
	assert.NotNil(t, err)
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "unable to list projects with search pattern"))
}

func readProjects(until chan struct{}, projects ...schemas.Project) <-chan schemas.Project {
	p := make(chan schemas.Project)
	go func() {
		defer close(p)
		for _, i := range projects {
			select {
			case <-until:
				return
			case p <- i:
			}
		}
	}()
	return p
}
