package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectTags(t *testing.T) {
	mux, server, c := getMockedClient()
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

	tags, err := c.GetProjectTags("foo", "[", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, tags, 0)

	tags, err = c.GetProjectTags("foo", "^f", 0)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, tags)
}

func TestGetProjectMostRecentTagCommit(t *testing.T) {
	mux, server, c := getMockedClient()
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

	_, _, err := c.GetProjectMostRecentTagCommit("foo", "[")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")

	commitShortID, commitCreatedAt, err := c.GetProjectMostRecentTagCommit("foo", "^f")
	assert.NoError(t, err)
	assert.Equal(t, "7b5c3cc", commitShortID)
	assert.Equal(t, float64(1553540113), commitCreatedAt)
}
