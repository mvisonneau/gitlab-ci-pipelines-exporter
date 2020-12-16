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
)

func TestGetRefPipeline(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"id":1}`)
		})

	ref := schemas.Ref{
		ProjectName: "foo",
		Name:        "yay",
	}

	pipeline, err := c.GetRefPipeline(ref, 1)
	assert.NoError(t, err)
	assert.NotNil(t, pipeline)
	assert.Equal(t, 1, pipeline.ID)
}

func TestGetProjectPipelines(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
				"ref":      []string{"foo"},
				"scope":    []string{"bar"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	pipelines, err := c.GetProjectPipelines("foo", &gitlab.ListProjectPipelinesOptions{
		Ref:   pointy.String("foo"),
		Scope: pointy.String("bar"),
	})

	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
}

func TestGetProjectMergeRequestsPipelines(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1,"ref":"refs/merge-requests/foo"},{"id":2,"ref":"refs/merge-requests/bar"},{"id":3,"ref":"yolo"}]`)
		})

	pipelines, err := c.GetProjectMergeRequestsPipelines("foo", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
}

func TestGetRefPipelineVariablesAsConcatenatedString(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"},{"key":"bar","value":"baz"}]`)
		})

	ref := schemas.Ref{
		ProjectName:                  "foo",
		Name:                         "yay",
		PullPipelineVariablesEnabled: true,
		PullPipelineVariablesRegexp:  "[",
	}

	// Should return right away as MostRecentPipeline is not defined
	variables, err := c.GetRefPipelineVariablesAsConcatenatedString(ref)
	assert.NoError(t, err)
	assert.Equal(t, "", variables)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should fail as we have an invalid regexp pattern
	variables, err = c.GetRefPipelineVariablesAsConcatenatedString(ref)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "the provided filter regex for pipeline variables is invalid")
	assert.Equal(t, "", variables)

	// Should work
	ref.PullPipelineVariablesRegexp = ".*"
	variables, err = c.GetRefPipelineVariablesAsConcatenatedString(ref)
	assert.NoError(t, err)
	assert.Equal(t, "foo:bar,bar:baz", variables)
}

func TestGetRefsFromPipelines(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			urlValues := r.URL.Query()
			assert.Equal(t, []string{"1"}, urlValues["page"])
			assert.Equal(t, []string{"100"}, urlValues["per_page"])

			if scope, ok := urlValues["scope"]; ok && len(scope) == 1 && scope[0] == "branches" {
				fmt.Fprint(w, `[{"id":1,"ref":"keep_dev"},{"id":2,"ref":"keep_main"}]`)
				return
			}

			if scope, ok := urlValues["scope"]; ok && len(scope) == 1 && scope[0] == "tags" {
				fmt.Fprint(w, `[{"id":3,"ref":"donotkeep_0.0.1"},{"id":4,"ref":"keep_0.0.2"}]`)
				return
			}

			fmt.Fprint(w, `{"error": "undefined or unsupported scope"`)
		})

	p := schemas.Project{
		Name: "foo",
		ProjectParameters: schemas.ProjectParameters{
			Pull: schemas.ProjectPull{
				Refs: schemas.ProjectPullRefs{
					RegexpValue: pointy.String("["), // invalid regexp pattern
					From: schemas.ProjectPullRefsFrom{
						Pipelines: schemas.ProjectPullRefsFromPipelines{
							EnabledValue: pointy.Bool(true),
							DepthValue:   pointy.Int(150),
						},
					},
				},
			},
		},
	}

	refs, err := c.GetRefsFromPipelines(p, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, refs, 0)

	p.Pull.Refs.RegexpValue = pointy.String("^keep.*")
	refs, err = c.GetRefsFromPipelines(p, "")
	assert.NoError(t, err)

	expectedRefs := schemas.Refs{
		"2231079763": schemas.Ref{
			Kind:                      schemas.RefKindBranch,
			ProjectName:               "foo",
			Name:                      "keep_dev",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
		"1035317703": schemas.Ref{
			Kind:                      schemas.RefKindBranch,
			ProjectName:               "foo",
			Name:                      "keep_main",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
		"1929034016": schemas.Ref{
			Kind:                      schemas.RefKindTag,
			ProjectName:               "foo",
			Name:                      "keep_0.0.2",
			LatestJobs:                make(schemas.Jobs),
			OutputSparseStatusMetrics: true,
			PullPipelineJobsFromChildPipelinesEnabled: true,
			PullPipelineVariablesRegexp:               ".*",
		},
	}

	assert.Equal(t, expectedRefs, refs)
}
