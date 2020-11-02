package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestGarbageCollectProjects(t *testing.T) {
	resetGlobalValues()

	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/groups/wc/projects",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1, "path_with_namespace": "wc/p3", "jobs_enabled": true}]`)
		})

	p1 := schemas.Project{Name: "cfg/p1"}
	p2 := schemas.Project{Name: "cfg/p2"}
	p3 := schemas.Project{Name: "wc/p3"}
	p4 := schemas.Project{Name: "wc/p4"}

	store.SetProject(p1)
	store.SetProject(p2)
	store.SetProject(p3)
	store.SetProject(p4)

	config = schemas.Config{
		Projects: []schemas.Project{p1},
		Wildcards: schemas.Wildcards{
			schemas.Wildcard{
				Owner: schemas.WildcardOwner{
					Kind: "group",
					Name: "wc",
				},
			},
		},
	}

	assert.NoError(t, garbageCollectProjects())
	storedProjects, err := store.Projects()
	assert.NoError(t, err)

	expectedProjects := schemas.Projects{
		p1.Key(): p1,
		p3.Key(): p3,
	}
	assert.Equal(t, expectedProjects, storedProjects)
}

func TestGarbageCollectEnvironments(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/p2/environments",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	p2 := schemas.Project{
		Name: "p2",
		ProjectParameters: schemas.ProjectParameters{
			Pull: schemas.ProjectPull{
				Environments: schemas.ProjectPullEnvironments{
					NameRegexpValue: pointy.String("^main$"),
				},
			},
		},
	}
	envp1main := schemas.Environment{ProjectName: "p1", Name: "main"}
	envp2dev := schemas.Environment{ProjectName: "p2", Name: "dev"}
	envp2main := schemas.Environment{ProjectName: "p2", Name: "main"}

	store.SetProject(p2)
	store.SetEnvironment(envp1main)
	store.SetEnvironment(envp2dev)
	store.SetEnvironment(envp2main)

	assert.NoError(t, garbageCollectEnvironments())
	storedEnvironments, err := store.Environments()
	assert.NoError(t, err)

	expectedEnvironments := schemas.Environments{
		envp2main.Key(): schemas.Environment{
			ProjectName:               "p2",
			Name:                      "main",
			TagsRegexp:                ".*",
			OutputSparseStatusMetrics: true,
		},
	}
	assert.Equal(t, expectedEnvironments, storedEnvironments)
}

func TestGarbageCollectRefs(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/p2/repository/branches",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	mux.HandleFunc("/api/v4/projects/p2/repository/tags",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name": "main"}]`)
		})

	pr1dev := schemas.Ref{Kind: schemas.RefKindBranch, ProjectName: "p1", Name: "dev"}
	pr1main := schemas.Ref{Kind: schemas.RefKindBranch, ProjectName: "p1", Name: "main"}

	p2 := schemas.Project{
		Name: "p2",
		ProjectParameters: schemas.ProjectParameters{
			Pull: schemas.ProjectPull{
				Refs: schemas.ProjectPullRefs{
					RegexpValue: pointy.String("^main$"),
				},
			},
		},
	}
	pr2dev := schemas.Ref{Kind: schemas.RefKindBranch, ProjectName: "p2", Name: "dev"}
	pr2main := schemas.Ref{Kind: schemas.RefKindBranch, ProjectName: "p2", Name: "main"}

	store.SetProject(p2)
	store.SetRef(pr1dev)
	store.SetRef(pr1main)
	store.SetRef(pr2dev)
	store.SetRef(pr2main)

	assert.NoError(t, garbageCollectRefs())
	storedRefs, err := store.Refs()
	assert.NoError(t, err)

	newPR2main := schemas.Ref{Kind: schemas.RefKindBranch, ProjectName: "p2", Name: "main"}
	expectedRefs := schemas.Refs{
		newPR2main.Key(): schemas.Ref{
			Kind:                        schemas.RefKindBranch,
			ProjectName:                 "p2",
			Name:                        "main",
			OutputSparseStatusMetrics:   true,
			PullPipelineVariablesRegexp: ".*",
		},
	}
	assert.Equal(t, expectedRefs, storedRefs)
}

func TestGarbageCollectMetrics(t *testing.T) {
	resetGlobalValues()

	ref1 := schemas.Ref{
		ProjectName:               "p1",
		Name:                      "foo",
		OutputSparseStatusMetrics: true,
		PullPipelineJobsEnabled:   true,
	}

	ref1m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}}
	ref1m2 := schemas.Metric{Kind: schemas.MetricKindStatus, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}}
	ref1m3 := schemas.Metric{Kind: schemas.MetricKindJobDurationSeconds, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}}

	ref2m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p2", "ref": "bar"}}
	ref3m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "foo"}}
	ref4m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"ref": "bar"}}

	store.SetRef(ref1)
	store.SetMetric(ref1m1)
	store.SetMetric(ref1m2)
	store.SetMetric(ref1m3)
	store.SetMetric(ref2m1)
	store.SetMetric(ref3m1)
	store.SetMetric(ref4m1)

	assert.NoError(t, garbageCollectMetrics())
	storedMetrics, err := store.Metrics()
	assert.NoError(t, err)

	expectedMetrics := schemas.Metrics{
		ref1m1.Key(): ref1m1,
		ref1m3.Key(): ref1m3,
	}
	assert.Equal(t, expectedMetrics, storedMetrics)
}
