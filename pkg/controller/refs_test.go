package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetRefs(t *testing.T) {
	c, mux, srv := newTestController(config.Config{})
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

	foundRefs, err := c.GetRefs(p)
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
	c, mux, srv := newTestController(config.Config{})
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
	assert.NoError(t, c.PullRefsFromProject(context.Background(), p1))

	ref1 := schemas.NewRef(p1, schemas.RefKindBranch, "main")
	expectedRefs := schemas.Refs{
		ref1.Key(): ref1,
	}

	projectsRefs, _ := c.Store.Refs()
	assert.Equal(t, expectedRefs, projectsRefs)
}

func TestPullRefsFromPipelines(t *testing.T) {
	c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"name":"foo"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "branches" {
				fmt.Fprint(w, `[{"id":1,"ref":"main"}]`)
				return
			}

			if scope, ok := r.URL.Query()["scope"]; ok && len(scope) == 1 && scope[0] == "tags" {
				fmt.Fprint(w, `[{"id":2,"ref":"v0.0.1"}]`)
				return
			}
		})

	p1 := schemas.NewProject("foo")
	p1.Pull.Refs.Branches.ExcludeDeleted = false
	p1.Pull.Refs.Tags.ExcludeDeleted = false

	assert.NoError(t, c.PullRefsFromPipelines(context.Background(), p1))

	ref1 := schemas.NewRef(p1, schemas.RefKindBranch, "main")
	ref2 := schemas.NewRef(p1, schemas.RefKindTag, "v0.0.1")
	expectedRefs := schemas.Refs{
		ref1.Key(): ref1,
		ref2.Key(): ref2,
	}

	projectsRefs, _ := c.Store.Refs()
	assert.Equal(t, expectedRefs, projectsRefs)
}
