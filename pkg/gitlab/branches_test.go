package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectBranches(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, []string{"100"}, r.URL.Query()["per_page"])
			currentPage, err := strconv.Atoi(r.URL.Query()["page"][0])
			assert.NoError(t, err)

			w.Header().Add("X-Total-Pages", "2")
			w.Header().Add("X-Page", strconv.Itoa(currentPage))
			w.Header().Add("X-Next-Page", strconv.Itoa(currentPage+1))

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

	branches, err := c.GetProjectBranches("1", "^(main)$", 0)
	assert.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Equal(t, "main", branches[0])

	// Test invalid project id
	_, err = c.GetProjectBranches("0", "", 0)
	assert.Error(t, err)

	// Test invalid regexp
	_, err = c.GetProjectBranches("0", "[", 0)
	assert.Error(t, err)
}

func TestGetBranchLatestCommit(t *testing.T) {
	mux, server, c := getMockedClient()
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

	commitShortID, commitCreatedAt, err := c.GetBranchLatestCommit("1", "main")
	assert.NoError(t, err)
	assert.Equal(t, "7b5c3cc", commitShortID)
	assert.Equal(t, float64(1553540113), commitCreatedAt)
}
