package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetProjectEnvironments(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(
		"/api/v4/projects/foo/environments",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, []string{"100"}, r.URL.Query()["per_page"])
			currentPage, err := strconv.Atoi(r.URL.Query()["page"][0])
			assert.NoError(t, err)
			nextPage := currentPage + 1
			if currentPage == 2 {
				nextPage = currentPage
			}

			w.Header().Add("X-Page", strconv.Itoa(currentPage))
			w.Header().Add("X-Next-Page", strconv.Itoa(nextPage))

			if scope, ok := r.URL.Query()["states"]; ok && len(scope) == 1 && scope[0] == "available" {
				fmt.Fprint(w, `[{"id":1338,"name":"main"}]`)

				return
			}

			if currentPage == 1 {
				fmt.Fprint(w, `[{"id":1338,"name":"main"},{"id":1337,"name":"dev"}]`)

				return
			}

			fmt.Fprint(w, `[]`)
		},
	)

	mux.HandleFunc(
		"/api/v4/projects/0/environments",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	p := schemas.NewProject("foo")
	p.Pull.Environments.Regexp = "^dev"
	p.Pull.Environments.ExcludeStopped = false

	xenv := schemas.Environment{
		ProjectName:               "foo",
		Name:                      "dev",
		ID:                        1337,
		OutputSparseStatusMetrics: true,
	}

	xenvs := schemas.Environments{
		xenv.Key(): xenv,
	}

	envs, err := c.GetProjectEnvironments(ctx, p)
	assert.NoError(t, err)
	assert.Equal(t, xenvs, envs)

	// Test invalid project
	p.Name = ""
	_, err = c.GetProjectEnvironments(ctx, p)
	assert.Error(t, err)

	// Test invalid regexp
	p.Name = "foo"
	p.Pull.Environments.Regexp = "["
	_, err = c.GetProjectEnvironments(ctx, p)
	assert.Error(t, err)

	// Test exclude stopped
	xenv = schemas.Environment{
		ProjectName:               "foo",
		Name:                      "main",
		ID:                        1338,
		OutputSparseStatusMetrics: true,
	}

	xenvs = schemas.Environments{
		xenv.Key(): xenv,
	}

	p.Pull.Environments.Regexp = ".*"
	p.Pull.Environments.ExcludeStopped = true
	envs, err = c.GetProjectEnvironments(ctx, p)
	assert.NoError(t, err)
	assert.Equal(t, xenvs, envs)
}

func TestGetEnvironment(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/environments/1",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `
{
	"id": 1,
	"name": "foo",
	"external_url": "https://foo.example.com",
	"state": "available",
	"last_deployment": {
		"ref": "bar",
		"created_at": "2019-03-25T18:55:13.252Z",
		"deployable": {
			"id": 23,
			"status": "success",
			"tag": false,
			"duration": 21623.13423,
			"user": {
				"username": "alice"
			},
			"commit": {
				"short_id": "416d8ea1"
			}
		}
	}
}`)
		})

	e, err := c.GetEnvironment(ctx, "foo", 1)
	assert.NoError(t, err)
	assert.NotNil(t, e)

	expectedEnv := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
		Name:        "foo",
		ExternalURL: "https://foo.example.com",
		Available:   true,
		LatestDeployment: schemas.Deployment{
			JobID:           23,
			RefKind:         schemas.RefKindBranch,
			RefName:         "bar",
			Username:        "alice",
			Timestamp:       1553540113,
			DurationSeconds: 21623.13423,
			CommitShortID:   "416d8ea1",
			Status:          "success",
		},
	}
	assert.Equal(t, expectedEnv, e)
}
