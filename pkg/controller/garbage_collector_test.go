package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

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
			fmt.Fprint(w, `[{"id":1, "path_with_namespace": "wc/p3", "jobs_enabled": true}]`)
		})

	c.Store.SetProject(ctx, p1)
	c.Store.SetProject(ctx, p2)
	c.Store.SetProject(ctx, p3)
	c.Store.SetProject(ctx, p4)

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
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	p2 := schemas.NewProject("p2")
	p2.Pull.Environments.Enabled = true
	p2.Pull.Environments.Regexp = "^main$"

	envp1main := schemas.Environment{ProjectName: "p1", Name: "main"}
	envp2dev := schemas.Environment{ProjectName: "p2", Name: "dev"}
	envp2main := schemas.Environment{ProjectName: "p2", Name: "main"}

	c.Store.SetProject(ctx, p2)
	c.Store.SetEnvironment(ctx, envp1main)
	c.Store.SetEnvironment(ctx, envp2dev)
	c.Store.SetEnvironment(ctx, envp2main)

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
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	mux.HandleFunc("/api/v4/projects/p2/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	pr1dev := schemas.NewRef(schemas.NewProject("p1"), schemas.RefKindBranch, "dev")
	pr1main := schemas.NewRef(schemas.NewProject("p1"), schemas.RefKindBranch, "main")

	p2 := schemas.NewProject("p2")
	p2.Pull.Environments.Regexp = "^main$"

	pr2dev := schemas.NewRef(p2, schemas.RefKindBranch, "dev")
	pr2main := schemas.NewRef(p2, schemas.RefKindBranch, "main")

	c.Store.SetProject(ctx, p2)
	c.Store.SetRef(ctx, pr1dev)
	c.Store.SetRef(ctx, pr1main)
	c.Store.SetRef(ctx, pr2dev)
	c.Store.SetRef(ctx, pr2main)

	assert.NoError(t, c.GarbageCollectRefs(context.Background()))
	storedRefs, err := c.Store.Refs(ctx)
	assert.NoError(t, err)

	newPR2main := schemas.NewRef(p2, schemas.RefKindBranch, "main")
	expectedRefs := schemas.Refs{
		newPR2main.Key(): newPR2main,
	}
	assert.Equal(t, expectedRefs, storedRefs)
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

	c.Store.SetRef(ctx, ref1)
	c.Store.SetMetric(ctx, ref1m1)
	c.Store.SetMetric(ctx, ref1m2)
	c.Store.SetMetric(ctx, ref1m3)
	c.Store.SetMetric(ctx, ref2m1)
	c.Store.SetMetric(ctx, ref3m1)
	c.Store.SetMetric(ctx, ref4m1)

	assert.NoError(t, c.GarbageCollectMetrics(context.Background()))
	storedMetrics, err := c.Store.Metrics(ctx)
	assert.NoError(t, err)

	expectedMetrics := schemas.Metrics{
		ref1m1.Key(): ref1m1,
		ref1m3.Key(): ref1m3,
	}
	assert.Equal(t, expectedMetrics, storedMetrics)
}
