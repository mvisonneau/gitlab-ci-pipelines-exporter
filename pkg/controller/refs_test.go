package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGetRefs(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/repository/branches",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"dev"},{"name":"main"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"0.0.1"},{"name":"v0.0.2"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"ref":"refs/merge-requests/1234/head"}]`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Refs.Branches.Regexp = `^m`
	p.Pull.Refs.Tags.Regexp = `^v`
	p.Pull.Refs.MergeRequests.Enabled = true

	foundRefs, err := c.GetRefs(ctx, p)
	assert.NoError(t, err)

	ref1 := schemas.NewRef(p, schemas.RefKindBranch, "main")
	ref2 := schemas.NewRef(p, schemas.RefKindTag, "v0.0.2")
	ref3 := schemas.NewRef(p, schemas.RefKindMergeRequest, "1234")
	expectedRefs := schemas.Refs{
		ref1.Key(): ref1,
		ref2.Key(): ref2,
		ref3.Key(): ref3,
	}
	assert.Equal(t, expectedRefs, foundRefs)
}

func TestPullRefsFromProject(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"name":"foo"}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/branches"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"main"},{"name":"nope"}]`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/repository/tags"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[]`)
		})

	p1 := schemas.NewProject("foo")
	assert.NoError(t, c.PullRefsFromProject(ctx, p1))

	ref1 := schemas.NewRef(p1, schemas.RefKindBranch, "main")
	expectedRefs := schemas.Refs{
		ref1.Key(): ref1,
	}

	projectsRefs, _ := c.Store.Refs(ctx)
	assert.Equal(t, expectedRefs, projectsRefs)
}
