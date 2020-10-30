package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

func TestGarbageCollectRefs(t *testing.T) {
	resetGlobalValues()

	pr1dev := schemas.Ref{PathWithNamespace: "p1", Ref: "dev"}
	pr1main := schemas.Ref{PathWithNamespace: "p1", Ref: "main"}

	p2old := schemas.Project{Name: "p2"}
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
	pr2dev := schemas.Ref{Project: p2old, PathWithNamespace: "p2", Ref: "dev"}
	pr2main := schemas.Ref{Project: p2old, PathWithNamespace: "p2", Ref: "main"}

	store.SetProject(p2)
	store.SetRef(pr1dev)
	store.SetRef(pr1main)
	store.SetRef(pr2dev)
	store.SetRef(pr2main)

	assert.NoError(t, garbageCollectRefs())
	storedRefs, err := store.Refs()
	assert.NoError(t, err)

	newPR2main := schemas.Ref{Project: p2, PathWithNamespace: "p2", Ref: "main"}
	expectedRefs := schemas.Refs{
		newPR2main.Key(): newPR2main,
	}
	assert.Equal(t, expectedRefs, storedRefs)
}

func TestGarbageCollectMetrics(t *testing.T) {
	resetGlobalValues()

	pr1 := schemas.Ref{
		Project: schemas.Project{
			ProjectParameters: schemas.ProjectParameters{
				OutputSparseStatusMetricsValue: pointy.Bool(true),
				Pull: schemas.ProjectPull{
					Pipeline: schemas.ProjectPullPipeline{
						Jobs: schemas.ProjectPullPipelineJobs{
							EnabledValue: pointy.Bool(false),
						},
					},
				},
			},
		},
		PathWithNamespace: "p1",
		Ref:               "foo",
	}

	pr1m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}}
	pr1m2 := schemas.Metric{Kind: schemas.MetricKindStatus, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}, Value: float64(0)}
	pr1m3 := schemas.Metric{Kind: schemas.MetricKindJobDurationSeconds, Labels: prometheus.Labels{"project": "p1", "ref": "foo"}}

	pr2m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "p2", "ref": "bar"}}
	pr3m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"project": "foo"}}
	pr4m1 := schemas.Metric{Kind: schemas.MetricKindCoverage, Labels: prometheus.Labels{"ref": "bar"}}

	store.SetRef(pr1)
	store.SetMetric(pr1m1)
	store.SetMetric(pr1m2)
	store.SetMetric(pr1m3)
	store.SetMetric(pr2m1)
	store.SetMetric(pr3m1)
	store.SetMetric(pr4m1)

	assert.NoError(t, garbageCollectMetrics())
	storedMetrics, err := store.Metrics()
	assert.NoError(t, err)

	expectedMetrics := schemas.Metrics{
		pr1m1.Key(): pr1m1,
	}
	assert.Equal(t, expectedMetrics, storedMetrics)
}

func TestMetricLogFields(t *testing.T) {
	m := schemas.Metric{
		Kind: schemas.MetricKindDurationSeconds,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	expected := log.Fields{
		"metric-kind":   schemas.MetricKindDurationSeconds,
		"metric-labels": prometheus.Labels{"foo": "bar"},
	}
	assert.Equal(t, expected, metricLogFields(m))
}

func TestStoreGetSetDelMetric(_ *testing.T) {
	resetGlobalValues()

	storeGetMetric(&schemas.Metric{})
	storeSetMetric(schemas.Metric{})
	storeDelMetric(schemas.Metric{})
}
