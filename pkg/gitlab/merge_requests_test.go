package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetProjectOpenMergeRequests(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/merge_requests",
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
				_, _ = fmt.Fprint(w, `[{"title":"Open MR", "iid":1}]`)

				return
			}

			_, _ = fmt.Fprint(w, `[]`)
		})

	mux.HandleFunc("/api/v4/projects/0/merge_requests",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

	p := schemas.NewProject("foo")
	expectedRef := schemas.NewRef(p, schemas.RefKindMergeRequest, "1")
	refs, err := c.GetProjectOpenMergeRequests(ctx, p)
	assert.NoError(t, err)
	assert.Len(t, refs, 1)
	assert.Equal(t, schemas.Refs{
		expectedRef.Key(): expectedRef,
	}, refs)

	// Test invalid project name
	p.Name = "invalid"
	_, err = c.GetProjectOpenMergeRequests(ctx, p)
	assert.Error(t, err)

	// Test invalid regexp
	p.Name = "foo"
	p.Pull.Refs.MergeRequests.Regexp = `[`
	_, err = c.GetProjectOpenMergeRequests(ctx, p)
	assert.Error(t, err)
}
