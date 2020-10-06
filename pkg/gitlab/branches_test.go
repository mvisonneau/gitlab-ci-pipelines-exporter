package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectBranches(t *testing.T) {
	mux, server, c := getMockedClient()
	defer server.Close()

	projectID := 1
	search := "^(main)$"
	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%d/repository/branches", projectID),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"20"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"name":"main"},{"name":"dev"}]`)
		})

	branches, err := c.GetProjectBranches(projectID, search)
	assert.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Equal(t, "main", branches[0])
}
