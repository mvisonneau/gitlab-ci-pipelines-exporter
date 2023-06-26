package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommitCountBetweenRefs(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/repository/compare",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"commits":[{},{},{}]}`)
		})

	mux.HandleFunc("/api/v4/projects/bar/repository/compare",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{`)
		})

	commitCount, err := c.GetCommitCountBetweenRefs(ctx, "foo", "bar", "baz")
	assert.NoError(t, err)
	assert.Equal(t, 3, commitCount)

	commitCount, err = c.GetCommitCountBetweenRefs(ctx, "bar", "", "")
	assert.Error(t, err)
	assert.Equal(t, 0, commitCount)
}
