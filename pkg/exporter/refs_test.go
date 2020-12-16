package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetRefs(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/repository/branches",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"keep/dev"},{"name":"keep/main"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"keep/dev"},{"name":"keep/0.0.2"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"ref":"refs/merge-requests/foo"}]`)
		})

	foundRefs, err := getRefs("foo", "^keep", 0, true, 10)
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

	mux.HandleFunc("/api/v4/projects/foo",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"name":"foo"}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"main"},{"name":"nope"}]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[]`)
		})

	assert.NoError(t, pullRefsFromProject(schemas.Project{Name: "foo"}))

	projectsRefs, _ := store.Refs()
	expectedRefs := schemas.Refs{
		"99908380": schemas.Ref{
			Kind:                      schemas.RefKindBranch,
			ProjectName:               "foo",
			Name:                      "main",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
	}
	assert.Equal(t, expectedRefs, projectsRefs)
}

func TestPullRefsFromPipelines(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"name":"foo"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
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

	assert.NoError(t, pullRefsFromPipelines(schemas.Project{Name: "foo"}))

	projectsRefs, _ := store.Refs()
	expectedRefs := schemas.Refs{
		"964648533": schemas.Ref{
			Kind:                      schemas.RefKindTag,
			ProjectName:               "foo",
			Name:                      "master",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
		"99908380": schemas.Ref{
			Kind:                      schemas.RefKindBranch,
			ProjectName:               "foo",
			Name:                      "main",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
	}
	assert.Equal(t, expectedRefs, projectsRefs)
}
