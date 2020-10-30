package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestGetRefs(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"keep/dev"},{"name":"keep/main"}]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"keep/dev"},{"name":"keep/0.0.2"}]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"ref":"refs/merge-requests/foo"}]`)
		})

	foundRefs, err := getRefs(1, "^keep", true, 10)
	assert.NoError(t, err)

	assert.Equal(t, foundRefs["keep/0.0.2"], schemas.RefKindTag)
	assert.Equal(t, foundRefs["keep/main"], schemas.RefKindBranch)
	assert.Equal(t, foundRefs["refs/merge-requests/foo"], schemas.RefKindMergeRequest)
	assert.Contains(t, []schemas.RefKind{schemas.RefKindTag, schemas.RefKindBranch}, foundRefs["keep/dev"])
}

func TestPullRefsFromProject(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/bar",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"id":1}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"main"},{"name":"nope"}]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[]`)
		})

	assert.NoError(t, pullRefsFromProject(schemas.Project{Name: "foo/bar"}))

	projectsRefs, _ := store.Refs()
	expectedRefs := schemas.Refs{
		"3207122276": schemas.Ref{
			Project: schemas.Project{
				Name: "foo/bar",
			},
			Kind: schemas.RefKindBranch,
			ID:   1,
			Ref:  "main",
			Jobs: make(map[string]goGitlab.Job),
		},
	}
	assert.Equal(t, expectedRefs, projectsRefs)
}

func TestPullRefsFromPipelines(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/bar",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"id":1}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "branches" {
				fmt.Fprint(w, `[{"id":1,"ref":"main"}]`)
				return
			}

			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "tags" {
				fmt.Fprint(w, `[{"id":2,"ref":"master"}]`)
				return
			}
		})

	assert.NoError(t, pullRefsFromPipelines(schemas.Project{Name: "foo/bar"}))

	projectsRefs, _ := store.Refs()
	expectedRefs := schemas.Refs{
		"3207122276": schemas.Ref{
			Project: schemas.Project{
				Name: "foo/bar",
			},
			Kind: schemas.RefKindBranch,
			ID:   1,
			Ref:  "main",
			Jobs: make(map[string]goGitlab.Job),
		},
		"755606486": schemas.Ref{
			Project: schemas.Project{
				Name: "foo/bar",
			},
			Kind: schemas.RefKindTag,
			ID:   1,
			Ref:  "master",
			Jobs: make(map[string]goGitlab.Job),
		},
	}
	assert.Equal(t, expectedRefs, projectsRefs)
}
