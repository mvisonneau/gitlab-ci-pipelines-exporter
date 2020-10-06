package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestGetProjectRefPipeline(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines/1"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"id":1}`)
		})

	pr := schemas.ProjectRef{
		ID:                1,
		PathWithNamespace: "foo/bar",
		Ref:               "yay",
	}

	pipeline, err := c.GetProjectRefPipeline(pr, 1)
	assert.NoError(t, err)
	assert.NotNil(t, pipeline)
	assert.Equal(t, 1, pipeline.ID)
}

func TestGetProjectPipelines(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"20"},
				"ref":      []string{"foo"},
				"scope":    []string{"bar"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	pipelines, err := c.GetProjectPipelines(1, &gitlab.ListProjectPipelinesOptions{
		Ref:   pointy.String("foo"),
		Scope: pointy.String("bar"),
	})

	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
}

func TestGetProjectMergeRequestsPipelines(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"20"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1,"ref":"refs/merge-requests/foo"},{"id":2,"ref":"refs/merge-requests/bar"},{"id":3,"ref":"yolo"}]`)
		})

	pipelines, err := c.GetProjectMergeRequestsPipelines(1, 10)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
}

func TestGetProjectRefPipelineVariablesAsConcatenatedString(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines/1/variables"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"},{"key":"bar","value":"baz"}]`)
		})

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "yay",
		Project: schemas.Project{
			Parameters: schemas.Parameters{
				PipelineVariablesRegexpValue: pointy.String("["), // invalid regexp pattern
			},
		},
	}

	// Should return right away as MostRecentPipeline is not defined
	variables, err := c.GetProjectRefPipelineVariablesAsConcatenatedString(pr)
	assert.NoError(t, err)
	assert.Equal(t, "", variables)

	pr.MostRecentPipeline = &goGitlab.Pipeline{
		ID: 1,
	}

	// Should fail as we have an invalid regexp pattern
	variables, err = c.GetProjectRefPipelineVariablesAsConcatenatedString(pr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "the provided filter regex for pipeline variables is invalid")
	assert.Equal(t, "", variables)

	// Should work
	pr.Parameters.PipelineVariablesRegexpValue = pointy.String(".*")
	variables, err = c.GetProjectRefPipelineVariablesAsConcatenatedString(pr)
	assert.NoError(t, err)
	assert.Equal(t, "foo:bar,bar:baz", variables)
}

func TestGetProjectRefsFromPipelines(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			urlValues := r.URL.Query()
			assert.Equal(t, []string{"1"}, urlValues["page"])
			assert.Equal(t, []string{"10"}, urlValues["per_page"])

			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "branches" {
				fmt.Fprint(w, `[{"id":1,"ref":"keep_dev"},{"id":2,"ref":"keep_main"}]`)
				return
			}

			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "tags" {
				fmt.Fprint(w, `[{"id":3,"ref":"donotkeep_0.0.1"},{"id":4,"ref":"keep_0.0.2"}]`)
				return
			}

			fmt.Fprint(w, `{"error": "undefined or unsupported scope"`)
		})

	p := schemas.Project{
		Name: "foo/bar",
		Parameters: schemas.Parameters{
			RefsRegexpValue: pointy.String("["), // invalid regexp pattern
		},
	}

	gp := &goGitlab.Project{
		ID:                1,
		PathWithNamespace: "foo/bar",
	}

	prs, err := c.GetProjectRefsFromPipelines(p, gp, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, prs, 0)

	p.RefsRegexpValue = pointy.String("^keep.*")
	prs, err = c.GetProjectRefsFromPipelines(p, gp, 10)
	assert.NoError(t, err)
	assert.Len(t, prs, 3)
}
