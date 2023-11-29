package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"go.openly.dev/pointy"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetRefPipeline(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"id":1}`)
		})

	ref := schemas.Ref{
		Project: schemas.NewProject("foo"),
		Name:    "yay",
	}

	pipeline, err := c.GetRefPipeline(ctx, ref, 1)
	assert.NoError(t, err)
	assert.NotNil(t, pipeline)
	assert.Equal(t, 1, pipeline.ID)
}

func TestGetProjectPipelines(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
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

	pipelines, _, err := c.GetProjectPipelines(ctx, "foo", &gitlab.ListProjectPipelinesOptions{
		Ref:   pointy.String("foo"),
		Scope: pointy.String("bar"),
	})

	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
}

func TestGetRefPipelineVariablesAsConcatenatedString(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"},{"key":"bar","value":"baz"}]`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.Variables.Regexp = `[`
	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	variables, err := c.GetRefPipelineVariablesAsConcatenatedString(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, "", variables)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should fail as we have an invalid regexp pattern
	variables, err = c.GetRefPipelineVariablesAsConcatenatedString(ctx, ref)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "the provided filter regex for pipeline variables is invalid")
	assert.Equal(t, "", variables)

	// Should work
	ref.Project.Pull.Pipeline.Variables.Regexp = `.*`
	variables, err = c.GetRefPipelineVariablesAsConcatenatedString(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, "foo:bar,bar:baz", variables)
}

func TestGetRefsFromPipelines(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()
	log.SetLevel(log.TraceLevel)

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"keep_main"}]`)

			return
		})

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

			fmt.Fprint(w, `[{"id":1,"ref":"keep_dev"},{"id":2,"ref":"keep_main"},{"id":3,"ref":"donotkeep_0.0.1"},{"id":4,"ref":"keep_0.0.2"},{"id":5,"ref":"refs/merge-requests/1234/head"}]`)
		})

	p := schemas.NewProject("foo")

	// Branches
	p.Pull.Refs.Branches.Regexp = `[` // invalid regexp pattern
	refs, err := c.GetRefsFromPipelines(ctx, p, schemas.RefKindBranch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, refs, 0)

	p.Pull.Refs.Branches.Regexp = "^keep.*"
	refs, err = c.GetRefsFromPipelines(ctx, p, schemas.RefKindBranch)
	assert.NoError(t, err)

	assert.Equal(t, schemas.Refs{
		"1035317703": schemas.NewRef(p, schemas.RefKindBranch, "keep_main"),
	}, refs)

	// Tags
	p.Pull.Refs.Tags.Regexp = `[` // invalid regexp pattern
	refs, err = c.GetRefsFromPipelines(ctx, p, schemas.RefKindTag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, refs, 0)

	p.Pull.Refs.Tags.Regexp = `^keep`
	p.Pull.Refs.Tags.ExcludeDeleted = false
	refs, err = c.GetRefsFromPipelines(ctx, p, schemas.RefKindTag)
	assert.NoError(t, err)

	assert.Equal(t, schemas.Refs{
		"1929034016": schemas.NewRef(p, schemas.RefKindTag, "keep_0.0.2"),
	}, refs)

	// Merge requests
	refs, err = c.GetRefsFromPipelines(ctx, p, schemas.RefKindMergeRequest)
	assert.NoError(t, err)
	assert.Equal(t, schemas.Refs{
		"622996356": schemas.NewRef(p, schemas.RefKindMergeRequest, "1234"),
	}, refs)
}

func TestGetRefPipelineTestReport(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	p := schemas.NewProject("foo")

	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	tr, err := c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{}, tr)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should work
	tr, err = c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{
		TotalTime:    5,
		TotalCount:   1,
		SuccessCount: 1,
		FailedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []schemas.TestSuite{
			{
				Name:         "Secure",
				TotalTime:    5,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 5,
						Status:        "success",
					},
				},
			},
		},
	}, tr)
}

func TestGetRefPipelineFailedTestReport(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 2, "success_count": 1, "failed_count": 1, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 2, "success_count": 1, "failed_count": 1, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "failed", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": "Failed message", "stack_trace": null}]}]}`)
		})

	p := schemas.NewProject("foo")

	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	tr, err := c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{}, tr)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should work
	tr, err = c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{
		TotalTime:    5,
		TotalCount:   2,
		SuccessCount: 1,
		FailedCount:  1,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []schemas.TestSuite{
			{
				Name:         "Secure",
				TotalTime:    5,
				TotalCount:   2,
				SuccessCount: 1,
				FailedCount:  1,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 5,
						Status:        "failed",
					},
				},
			},
		},
	}, tr)
}

func TestGetRefPipelineWithParentChildTestReport(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/bridges",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"id":1,"downstream_pipeline":{"id":2, "project_id": 1}}]`)
		})

	mux.HandleFunc("/api/v4/projects/1/pipelines/2/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 3, "total_count": 3, "success_count": 3, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 3, "total_count": 3, "success_count": 3, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 3, "system_output": null, "stack_trace": null}]}]}`)
		})

	mux.HandleFunc("/api/v4/projects/1/pipelines/2/bridges",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[]`)
		})

	p := schemas.NewProject("foo")

	p.Project.Pull.Pipeline.TestReports.FromChildPipelines.Enabled = true

	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	tr, err := c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{}, tr)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should work
	tr, err = c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{
		TotalTime:    8,
		TotalCount:   4,
		SuccessCount: 4,
		FailedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []schemas.TestSuite{
			{
				Name:         "Secure",
				TotalTime:    5,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 5,
						Status:        "success",
					},
				},
			},
			{
				Name:         "Secure",
				TotalTime:    3,
				TotalCount:   3,
				SuccessCount: 3,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 3,
						Status:        "success",
					},
				},
			},
		},
	}, tr)
}

func TestGetRefPipelineWithMultiProjectTestReport(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/bridges",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"id":1,"downstream_pipeline":{"id":2, "project_id": 11}}]`)
		})

	mux.HandleFunc("/api/v4/projects/11/pipelines/2/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 3, "total_count": 3, "success_count": 3, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 3, "total_count": 3, "success_count": 3, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 3, "system_output": null, "stack_trace": null}]}]}`)
		})

	mux.HandleFunc("/api/v4/projects/11/pipelines/2/bridges",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[]`)
		})

	p := schemas.NewProject("foo")

	p.Project.Pull.Pipeline.TestReports.FromChildPipelines.Enabled = true

	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	tr, err := c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{}, tr)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should work
	tr, err = c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{
		TotalTime:    8,
		TotalCount:   4,
		SuccessCount: 4,
		FailedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []schemas.TestSuite{
			{
				Name:         "Secure",
				TotalTime:    5,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 5,
						Status:        "success",
					},
				},
			},
			{
				Name:         "Secure",
				TotalTime:    3,
				TotalCount:   3,
				SuccessCount: 3,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 3,
						Status:        "success",
					},
				},
			},
		},
	}, tr)
}

func TestGetRefPipelineWithNoChildrenTestReport(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/bridges",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[]`)
		})

	p := schemas.NewProject("foo")

	p.Project.Pull.Pipeline.TestReports.FromChildPipelines.Enabled = true

	ref := schemas.Ref{
		Project: p,
		Name:    "yay",
	}

	// Should return right away as MostRecentPipeline is not defined
	tr, err := c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{}, tr)

	ref.LatestPipeline = schemas.Pipeline{
		ID: 1,
	}

	// Should work
	tr, err = c.GetRefPipelineTestReport(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, schemas.TestReport{
		TotalTime:    5,
		TotalCount:   1,
		SuccessCount: 1,
		FailedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []schemas.TestSuite{
			{
				Name:         "Secure",
				TotalTime:    5,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []schemas.TestCase{
					{
						Name:          "Security Reports can create an auto-remediation MR",
						Classname:     "vulnerability_management_spec",
						ExecutionTime: 5,
						Status:        "success",
					},
				},
			},
		},
	}, tr)
}
