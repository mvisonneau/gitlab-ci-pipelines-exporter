package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestListProjectRefPipelineJobs(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "yay",
	}

	jobs, err := c.ListProjectRefPipelineJobs(pr)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines/1/jobs"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		})

	pr.MostRecentPipeline = &goGitlab.Pipeline{
		ID: 1,
	}

	jobs, err = c.ListProjectRefPipelineJobs(pr)
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
}

func TestListProjectRefMostRecentJobs(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "yay",
	}

	jobs, err := c.ListProjectRefMostRecentJobs(pr)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/jobs"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":3,"name":"foo","ref":"yay"},{"id":4,"name":"bar","ref":"yay"}]`)
		})

	pr.Jobs = map[string]goGitlab.Job{
		"foo": {
			ID:   1,
			Name: "foo",
		},
		"bar": {
			ID:   2,
			Name: "bar",
		},
	}

	jobs, err = c.ListProjectRefMostRecentJobs(pr)
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
	assert.Equal(t, 3, jobs[0].ID)
	assert.Equal(t, 4, jobs[1].ID)

	pr.Jobs["baz"] = goGitlab.Job{
		ID:   5,
		Name: "baz",
	}

	jobs, err = c.ListProjectRefMostRecentJobs(pr)
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
	assert.Equal(t, 3, jobs[0].ID)
	assert.Equal(t, 4, jobs[1].ID)
}
