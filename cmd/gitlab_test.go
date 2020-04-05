package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gitlab "github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

// Mocking helpers
func getMockedGitlabClient() (*http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	opts := []gitlab.ClientOptionFunc{
		gitlab.WithBaseURL(server.URL),
		gitlab.WithoutRetries(),
	}

	gc, _ := gitlab.NewClient("", opts...)

	c := &Client{
		Client:      gc,
		RateLimiter: ratelimit.New(100),
	}

	return mux, server, c
}

// Functions testing
func TestProjectExists(t *testing.T) {
	foo := Project{Name: "foo"}
	bar := Project{Name: "bar"}

	cfg = &Config{
		Projects: []Project{foo},
	}

	assert.Equal(t, true, projectExists(foo))
	assert.Equal(t, false, projectExists(bar))
}

func TestRefExists(t *testing.T) {
	refs := []string{"foo"}

	assert.Equal(t, true, refExists(refs, "foo"))
	assert.Equal(t, false, refExists(refs, "bar"))
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
	assert.Equal(t, 1, p.ID)
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
	assert.Equal(t, 2, len(projects))
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
	assert.Equal(t, 2, len(projects))
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
	assert.Equal(t, 2, len(projects))
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
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "Unable to list projects with search pattern"))
}
