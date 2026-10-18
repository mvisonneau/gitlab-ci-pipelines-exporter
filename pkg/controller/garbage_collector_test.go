package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestGarbageCollectProjects(t *testing.T) {
	p1 := schemas.NewProject("cfg/p1")
	p2 := schemas.NewProject("cfg/p2")
	p3 := schemas.NewProject("wc/p3")
	p4 := schemas.NewProject("wc/p4")

	ctx, c, mux, srv := newTestController(config.Config{
		Projects: []config.Project{p1.Project},
		Wildcards: config.Wildcards{
			config.Wildcard{
				Owner: config.WildcardOwner{
					Kind: "group",
					Name: "wc",
				},
			},
		},
	})
	defer srv.Close()

	mux.HandleFunc("/api/v4/groups/wc/projects",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `[{"id":1, "path_with_namespace": "wc/p3", "jobs_enabled": true}]`)
		})

	_ = c.Store.SetProject(ctx, p1)
	_ = c.Store.SetProject(ctx, p2)
	_ = c.Store.SetProject(ctx, p3)
	_ = c.Store.SetProject(ctx, p4)

	assert.NoError(t, c.GarbageCollectProjects(context.Background()))
	storedProjects, err := c.Store.Projects(ctx)
	assert.NoError(t, err)

	expectedProjects := schemas.Projects{
		p1.Key(): p1,
		p3.Key(): p3,
	}
	assert.Equal(t, expectedProjects, storedProjects)
}

func TestGarbageCollectEnvironments(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/p2/environments",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `[{"name": "main"}]`)
		})

	p2 := schemas.NewProject("p2")
	p2.Pull.Environments.Enabled = true
	p2.Pull.Environments.Regexp = "^main$"

	envp1main := schemas.Environment{ProjectName: "p1", Name: "main"}
	envp2dev := schemas.Environment{ProjectName: "p2", Name: "dev"}
	envp2main := schemas.Environment{ProjectName: "p2", Name: "main"}

	_ = c.Store.SetProject(ctx, p2)
	_ = c.Store.SetEnvironment(ctx, envp1main)
	_ = c.Store.SetEnvironment(ctx, envp2dev)
	_ = c.Store.SetEnvironment(ctx, envp2main)

	assert.NoError(t, c.GarbageCollectEnvironments(context.Background()))
	storedEnvironments, err := c.Store.Environments(ctx)
	assert.NoError(t, err)

	expectedEnvironments := schemas.Environments{
		envp2main.Key(): schemas.Environment{
			ProjectName:               "p2",
			Name:                      "main",
			OutputSparseStatusMetrics: true,
		},
	}
	assert.Equal(t, expectedEnvironments, storedEnvironments)
}

func TestGarbageCollectRefs(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/p2/repository/branches",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `[{"name": "main"}]`)
		})

	mux.HandleFunc("/api/v4/projects/p2/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `[{"name": "main"}]`)
		})

	pr1dev := schemas.NewRef(schemas.NewProject("p1"), schemas.RefKindBranch, "dev")
	pr1main := schemas.NewRef(schemas.NewProject("p1"), schemas.RefKindBranch, "main")

	p2 := schemas.NewProject("p2")
	p2.Pull.Environments.Regexp = "^main$"

	pr2dev := schemas.NewRef(p2, schemas.RefKindBranch, "dev")
	pr2main := schemas.NewRef(p2, schemas.RefKindBranch, "main")

	_ = c.Store.SetProject(ctx, p2)
	_ = c.Store.SetRef(ctx, pr1dev)
	_ = c.Store.SetRef(ctx, pr1main)
	_ = c.Store.SetRef(ctx, pr2dev)
	_ = c.Store.SetRef(ctx, pr2main)

	assert.NoError(t, c.GarbageCollectRefs(context.Background()))
	storedRefs, err := c.Store.Refs(ctx)
	assert.NoError(t, err)

	newPR2main := schemas.NewRef(p2, schemas.RefKindBranch, "main")
	expectedRefs := schemas.Refs{
		newPR2main.Key(): newPR2main,
	}
	assert.Equal(t, expectedRefs, storedRefs)
}

func TestGarbageCollectRefsProjectFailureIsolation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		partial bool
	}{
		{name: "not found", status: http.StatusNotFound},
		{name: "forbidden", status: http.StatusForbidden},
		{name: "rate limited", status: http.StatusTooManyRequests},
		{name: "server error", status: http.StatusInternalServerError},
		{name: "partial discovery", status: http.StatusInternalServerError, partial: true},
		{name: "empty success", status: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, c, mux, srv := newTestController(config.Config{})
			defer srv.Close()

			healthy := schemas.NewProject("healthy")
			healthy.Pull.Refs.Branches.Regexp = ".*"
			healthy.Pull.Refs.Tags.Enabled = false
			failing := schemas.NewProject("failing")
			failing.Pull.Refs.Branches.Regexp = ".*"
			failing.Pull.Refs.Tags.Enabled = tc.partial
			empty := schemas.NewProject("empty")
			empty.Pull.Refs.Tags.Enabled = false

			for _, project := range []schemas.Project{healthy, failing, empty} {
				require.NoError(t, c.Store.SetProject(ctx, project))
			}

			failedBranch := schemas.NewRef(failing, schemas.RefKindBranch, "main")
			failedTag := schemas.NewRef(failing, schemas.RefKindTag, "v1")
			for _, ref := range []schemas.Ref{
				failedBranch, failedTag,
				schemas.NewRef(empty, schemas.RefKindBranch, "main"),
				schemas.NewRef(healthy, schemas.RefKindBranch, "obsolete"),
			} {
				require.NoError(t, c.Store.SetRef(ctx, ref))
			}

			var currentBranch string
			mux.HandleFunc("/api/v4/projects/healthy/repository/branches", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `[{"name": %q}]`, currentBranch)
			})
			mux.HandleFunc("/api/v4/projects/empty/repository/branches", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, `[]`)
			})
			mux.HandleFunc("/api/v4/projects/failing/repository/branches", func(w http.ResponseWriter, _ *http.Request) {
				if tc.partial {
					_, _ = fmt.Fprint(w, `[{"name": "main"}]`)
					return
				}
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, `[]`)
			})
			mux.HandleFunc("/api/v4/projects/failing/repository/tags", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, `[]`)
			})

			// Simulate discovery adding refs between successive GC runs.
			for cycle := range 3 {
				currentBranch = fmt.Sprintf("branch-%d", cycle)
				currentRef := schemas.NewRef(healthy, schemas.RefKindBranch, currentBranch)
				require.NoError(t, c.Store.SetRef(ctx, currentRef))

				err := c.GarbageCollectRefs(ctx)
				expected := schemas.Refs{currentRef.Key(): currentRef}
				if tc.status == http.StatusOK {
					assert.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(t, err.Error(), failing.Name)
					assert.Contains(t, err.Error(), fmt.Sprint(tc.status))
					expected[failedBranch.Key()] = failedBranch
					expected[failedTag.Key()] = failedTag
				}

				storedRefs, err := c.Store.Refs(ctx)
				require.NoError(t, err)
				assert.Equal(t, expected, storedRefs)
				projects, err := c.Store.Projects(ctx)
				require.NoError(t, err)
				assert.Contains(t, projects, failing.Key())
			}
		})
	}
}

func TestGarbageCollectMetrics(t *testing.T) {
	ctx, c, _, srv := newTestController(config.Config{})
	srv.Close()

	p1 := schemas.NewProject("p1")
	p1.Pull.Pipeline.Jobs.Enabled = true

	ref1 := schemas.NewRef(p1, schemas.RefKindBranch, "foo")

	ref1m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p1", "ref": "foo", "kind": "branch"}}
	ref1m2 := schemas.Metric{Kind: schemas.MetricKindStatus, Labels: prometheus.Labels{"project": "p1", "ref": "foo", "kind": "branch"}}
	ref1m3 := schemas.Metric{Kind: schemas.MetricKindJobDurationSeconds, Labels: prometheus.Labels{"project": "p1", "ref": "foo", "kind": "branch"}}

	ref2m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p2", "ref": "bar", "kind": "branch"}}
	ref3m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "foo", "kind": "branch"}}
	ref4m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"ref": "bar", "kind": "branch"}}

	_ = c.Store.SetRef(ctx, ref1)
	_ = c.Store.SetMetric(ctx, ref1m1)
	_ = c.Store.SetMetric(ctx, ref1m2)
	_ = c.Store.SetMetric(ctx, ref1m3)
	_ = c.Store.SetMetric(ctx, ref2m1)
	_ = c.Store.SetMetric(ctx, ref3m1)
	_ = c.Store.SetMetric(ctx, ref4m1)

	assert.NoError(t, c.GarbageCollectMetrics(context.Background()))
	storedMetrics, err := c.Store.Metrics(ctx)
	assert.NoError(t, err)

	expectedMetrics := schemas.Metrics{
		ref1m1.Key(): ref1m1,
		ref1m3.Key(): ref1m3,
	}
	assert.Equal(t, expectedMetrics, storedMetrics)
}
