package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetProjectTags(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"name":"foo"},{"name":"bar"}]`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Refs.Tags.Regexp = `^f`

	expectedRef := schemas.NewRef(p, schemas.RefKindTag, "foo")
	refs, err := c.GetProjectTags(ctx, p)
	assert.NoError(t, err)
	assert.Len(t, refs, 1)
	assert.Equal(t, schemas.Refs{
		expectedRef.Key(): expectedRef,
	}, refs)

	// Test invalid project name
	p.Name = "invalid"
	_, err = c.GetProjectTags(ctx, p)
	assert.Error(t, err)

	// Test invalid regexp
	p.Name = "foo"
	p.Pull.Refs.Tags.Regexp = `[`
	_, err = c.GetProjectTags(ctx, p)
	assert.Error(t, err)
}

func TestGetProjectMostRecentTagCommit(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"100"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `
[
	{
		"name": "foo",
		"commit": {
			"short_id": "7b5c3cc",
			"committed_date": "2019-03-25T18:55:13.252Z"
		}
	},
	{
		"name": "bar"
	}
]`)
		})

	_, _, err := c.GetProjectMostRecentTagCommit(ctx, "foo", "[")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")

	commitShortID, commitCreatedAt, err := c.GetProjectMostRecentTagCommit(ctx, "foo", "^f")
	assert.NoError(t, err)
	assert.Equal(t, "7b5c3cc", commitShortID)
	assert.Equal(t, float64(1553540113), commitCreatedAt)
}
