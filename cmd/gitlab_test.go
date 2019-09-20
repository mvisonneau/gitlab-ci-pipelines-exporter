package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xanzy/go-gitlab"
)

// Mocking helpers
func getMockedGitlabClient() (*http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	c := &Client{gitlab.NewClient(&http.Client{}, "")}
	c.SetBaseURL(server.URL)

	return mux, server, c
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

// Functions testing
func TestProjectExists(t *testing.T) {
	foo := Project{Name: "foo"}
	bar := Project{Name: "bar"}

	cfg = &Config{
		Projects: []Project{foo},
	}

	if !projectExists(foo) {
		t.Fatalf("Expected project foo to exist")
	}

	if projectExists(bar) {
		t.Fatalf("Expected project bar to not exist")
	}
}

func TestRefExists(t *testing.T) {
	refs := []string{"foo"}

	if !refExists(refs, "foo") {
		t.Fatalf("Expected ref foo to exist")
	}

	if refExists(refs, "bar") {
		t.Fatalf("Expected ref bar to not exist")
	}
}

func TestGetProject(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	project := "foo/bar"
	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%s", project),
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `{"id":1}`)
		})

	p, err := c.getProject(project)
	if err != nil {
		t.Fatalf("Did not expect this error %v", err)
	}

	if p == nil {
		t.Fatalf("Expected 'p' to be defined, got nil")
	}

	if p.ID != 1 {
		t.Fatalf("Expected 'p.ID' to equal 1, got %d", p.ID)
	}
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
		Refs: "^master|1.0$",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	if err != nil {
		t.Fatalf("Did not expect this error %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected to get 2 projects, got %d", len(projects))
	}
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
		Refs: "^master|1.0$",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/groups/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	if err != nil {
		t.Fatalf("Did not expect this error %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected to get 2 projects, got %d", len(projects))
	}
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
		Refs: "",
	}

	mux.HandleFunc("/api/v4/projects",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	projects, err := c.listProjects(w)
	if err != nil {
		t.Fatalf("Did not expect this error %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected to get 2 projects, got %d", len(projects))
	}
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
		Refs: "",
	}

	mux.HandleFunc(fmt.Sprintf("/api/v4/users/%s/projects", w.Owner.Name),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		})

	_, err := c.listProjects(w)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !strings.HasPrefix(err.Error(), "Unable to list projects with search pattern") {
		t.Fatalf("Did not expect this error %v", err)
	}
}
