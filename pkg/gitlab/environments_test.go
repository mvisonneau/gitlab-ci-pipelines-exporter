package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetProjectEnvironments(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/environments"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, []string{"100"}, r.URL.Query()["per_page"])
			currentPage, err := strconv.Atoi(r.URL.Query()["page"][0])
			assert.NoError(t, err)

			w.Header().Add("X-Total-Pages", "2")
			w.Header().Add("X-Page", strconv.Itoa(currentPage))
			w.Header().Add("X-Next-Page", strconv.Itoa(currentPage+1))

			if currentPage == 1 {
				fmt.Fprint(w, `[{"name":"main"},{"id":1337,"name":"dev"}]`)
				return
			}

			fmt.Fprint(w, `[]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/0/environments"),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

	envs, err := c.GetProjectEnvironments("foo", "^dev")
	assert.NoError(t, err)
	assert.Equal(t, map[int]string{1337: "dev"}, envs)

	// Test invalid project
	_, err = c.GetProjectEnvironments("0", "")
	assert.Error(t, err)

	// Test invalid regexp
	_, err = c.GetProjectEnvironments("1", "[")
	assert.Error(t, err)
}

func TestGetEnvironment(t *testing.T) {
	mux, server, c := getMockedClient()
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
				"public_email": "foo@bar.com"
			},
			"commit": {
				"short_id": "416d8ea1"
			}
		}
	}
}`)
		})

	e, err := c.GetEnvironment("foo", 1)
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
			AuthorEmail:     "foo@bar.com",
			Timestamp:       1553540113,
			DurationSeconds: 21623.13423,
			CommitShortID:   "416d8ea1",
			Status:          "success",
		},
	}
	assert.Equal(t, expectedEnv, e)
}
