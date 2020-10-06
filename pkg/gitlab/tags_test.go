package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectTags(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{
				"page":     []string{"1"},
				"per_page": []string{"20"},
			}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"name":"foo"},{"name":"bar"}]`)
		})

	tags, err := c.GetProjectTags(1, "[")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
	assert.Len(t, tags, 0)

	tags, err = c.GetProjectTags(1, "^f")
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, tags)
}
