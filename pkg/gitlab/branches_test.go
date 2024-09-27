package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetProjectBranches(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/branches"),
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

			if currentPage == 1 {
				fmt.Fprint(w, `[{"name":"main"},{"name":"dev"}]`)

				return
			}

			fmt.Fprint(w, `[]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/0/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

	p := schemas.NewProject("foo")
	expectedRef := schemas.NewRef(p, schemas.RefKindBranch, "main")
	refs, err := c.GetProjectBranches(ctx, p)
	assert.NoError(t, err)
	assert.Len(t, refs, 1)
	assert.Equal(t, schemas.Refs{
		expectedRef.Key(): expectedRef,
	}, refs)

	// Test invalid project name
	p.Name = "invalid"
	_, err = c.GetProjectBranches(ctx, p)
	assert.Error(t, err)

	// Test invalid regexp
	p.Name = "foo"
	p.Pull.Refs.Branches.Regexp = `[`
	_, err = c.GetProjectBranches(ctx, p)
	assert.Error(t, err)
}

func TestGetBranchLatestCommit(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/1/repository/branches/main",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `
{
	"commit": {
		"short_id": "7b5c3cc",
		"committed_date": "2019-03-25T18:55:13.252Z"
	}
}`)
		})

	commitShortID, commitCreatedAt, err := c.GetBranchLatestCommit(ctx, "1", "main")
	assert.NoError(t, err)
	assert.Equal(t, "7b5c3cc", commitShortID)
	assert.Equal(t, float64(1553540113), commitCreatedAt)
}
