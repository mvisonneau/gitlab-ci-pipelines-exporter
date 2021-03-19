package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetProject(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	project := "foo/bar"
	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%s", project),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `{"id":1}`)
		})

	p, err := c.GetProject(project)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, 1, p.ID)
}

func TestListUserProjects(t *testing.T) {
	mux, server, c := getMockedClient()
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
			fmt.Fprint(w, `[{"id":1,"path_with_namespace":"foo/bar","jobs_enabled":true},{"id":2,"path_with_namespace":"bar/baz","jobs_enabled":true}]`)
		})

	projects, err := c.ListProjects(w)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "foo/bar", projects[0].Name)
}

func TestListGroupProjects(t *testing.T) {
	mux, server, c := getMockedClient()
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
			fmt.Fprint(w, `[{"id":1,"path_with_namespace":"foo/bar","jobs_enabled":true},{"id":2,"path_with_namespace":"bar/baz","jobs_enabled":true}]`)
		})

	projects, err := c.ListProjects(w)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "foo/bar", projects[0].Name)
}

func TestListProjects(t *testing.T) {
	mux, server, c := getMockedClient()
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
			fmt.Fprint(w, `[{"id":1,"path_with_namespace":"foo","jobs_enabled":false},{"id":2,"path_with_namespace":"bar","jobs_enabled":true}]`)
		})

	projects, err := c.ListProjects(w)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "bar", projects[0].Name)
}

func TestListProjectsAPIError(t *testing.T) {
	mux, server, c := getMockedClient()
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
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to list projects with search pattern")
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
