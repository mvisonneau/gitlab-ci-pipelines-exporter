package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

// Mocking helpers
func getMockedGitlabClient() (*http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	c := &Client{
		Client:      gitlab.NewClient(&http.Client{}, ""),
		RateLimiter: ratelimit.New(100),
	}
	c.SetBaseURL(server.URL)

	return mux, server, c
}

// Functions testing
func TestProjectExists(t *testing.T) {
	foo := Project{Name: "foo"}
	bar := Project{Name: "bar"}

	cfg = &Config{
		Projects: []Project{foo},
	}

	assert.Equal(t, projectExists(foo), true)
	assert.Equal(t, projectExists(bar), false)
}

func TestRefExists(t *testing.T) {
	refs := []string{"foo"}

	assert.Equal(t, refExists(refs, "foo"), true)
	assert.Equal(t, refExists(refs, "bar"), false)
}

func TestGetProject(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	project := "foo/bar"
	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%s", project),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `{"id":1}`)
		})

	p, err := c.getProject(project)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, p.ID, 1)
}

func TestListUserProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := &Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string
			Kind             string
			IncludeSubgroups bool `yaml:"include_subgroups"`
		}{
			Name:             "foo",
			Kind:             "user",
			IncludeSubgroups: false,
		},
		Archived: false,
		Refs:     "^master|1.0$",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, len(projects), 2)
}

func TestListGroupProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := &Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string
			Kind             string
			IncludeSubgroups bool `yaml:"include_subgroups"`
		}{
			Name:             "foo",
			Kind:             "group",
			IncludeSubgroups: false,
		},
		Archived: false,
		Refs:     "^master|1.0$",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/groups/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, len(projects), 2)
}

func TestListProjects(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := &Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string
			Kind             string
			IncludeSubgroups bool `yaml:"include_subgroups"`
		}{
			Name:             "",
			Kind:             "",
			IncludeSubgroups: false,
		},
		Archived: false,
		Refs:     "",
	}

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	assert.Nil(t, err)
	assert.Equal(t, len(projects), 2)
}

func TestListProjectsAPIError(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	w := &Wildcard{
		Search: "bar",
		Owner: struct {
			Name             string
			Kind             string
			IncludeSubgroups bool `yaml:"include_subgroups"`
		}{
			Name: "foo",
			Kind: "user",
		},
		Archived: false,
		Refs:     "",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		})

	_, err := c.listProjects(w)
	assert.NotNil(t, err)
	assert.Equal(t, strings.HasPrefix(err.Error(), "Unable to list projects with search pattern"), true)
}
