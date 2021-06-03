package controller

import (
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestTriggerRefMetricsPull(_ *testing.T) {
	c, _, srv := newTestController(config.Config{})
	srv.Close()

	ref1 := schemas.Ref{
		Project: schemas.NewProject("group/foo"),
		Name:    "main",
	}

	p2 := schemas.NewProject("group/bar")
	ref2 := schemas.Ref{
		Project: p2,
		Name:    "main",
	}

	c.Store.SetRef(ref1)
	c.Store.SetProject(p2)

	// TODO: Assert results somehow
	c.triggerRefMetricsPull(ref1)
	c.triggerRefMetricsPull(ref2)
}

func TestTriggerEnvironmentMetricsPull(_ *testing.T) {
	c, _, srv := newTestController(config.Config{})
	srv.Close()

	p1 := schemas.NewProject("foo/bar")
	env1 := schemas.Environment{
		ProjectName: p1.Name,
		Name:        "dev",
	}

	env2 := schemas.Environment{
		ProjectName: "foo/baz",
		Name:        "prod",
	}

	c.Store.SetProject(p1)
	c.Store.SetEnvironment(env1)
	c.Store.SetEnvironment(env2)

	// TODO: Assert results somehow
	c.triggerEnvironmentMetricsPull(env1)
	c.triggerEnvironmentMetricsPull(env2)
}
