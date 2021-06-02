package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestPullRefPipelineJobsMetrics(t *testing.T) {
	c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	ref := schemas.Ref{
		ProjectName: "foo",
		Name:        "bar",
		LatestPipeline: schemas.Pipeline{
			ID: 1,
		},
	}

	assert.NoError(t, c.PullRefPipelineJobsMetrics(ref))
	srv.Close()
	assert.Error(t, c.PullRefPipelineJobsMetrics(ref))
}

func TestPullRefMostRecentJobsMetrics(t *testing.T) {
	c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	ref := schemas.Ref{
		ProjectName: "foo",
		Name:        "bar",
		LatestJobs: schemas.Jobs{
			"bar": {
				ID: 1,
			},
		},
	}

	// Test with FetchPipelineJobMetrics disabled
	assert.NoError(t, c.PullRefMostRecentJobsMetrics(ref))

	// Enable FetchPipelineJobMetrics
	ref.PullPipelineJobsEnabled = true
	assert.NoError(t, c.PullRefMostRecentJobsMetrics(ref))
	srv.Close()
	assert.Error(t, c.PullRefMostRecentJobsMetrics(ref))
}

func TestProcessJobMetrics(t *testing.T) {
	c, _, srv := newTestController(config.Config{})
	srv.Close()

	oldJob := schemas.Job{
		ID:        1,
		Name:      "foo",
		Timestamp: 1,
	}

	newJob := schemas.Job{
		ID:              2,
		Name:            "foo",
		Timestamp:       2,
		DurationSeconds: 15,
		Status:          "failed",
		Stage:           "🚀",
		ArtifactSize:    150,
		Runner: schemas.Runner{
			Description: "foo-123-bar",
		},
	}

	ref := schemas.Ref{
		ProjectName: "foo/bar",
		Topics:      "first,second",
		Kind:        schemas.RefKindBranch,
		Name:        "foo",
		LatestPipeline: schemas.Pipeline{
			ID:        1,
			Variables: "none",
		},
		LatestJobs: schemas.Jobs{
			"foo": oldJob,
		},
		OutputSparseStatusMetrics:                          true,
		PullPipelineJobsRunnerDescriptionEnabled:           true,
		PullPipelineJobsRunnerDescriptionAggregationRegexp: "foo-(.*)-bar",
	}

	c.Store.SetRef(ref)

	// If we run it against the same job, nothing should change in the store
	c.ProcessJobMetrics(ref, oldJob)
	refs, _ := c.Store.Refs()
	assert.Equal(t, schemas.Jobs{
		"foo": oldJob,
	}, refs[ref.Key()].LatestJobs)

	// Update the ref
	c.ProcessJobMetrics(ref, newJob)
	refs, _ = c.Store.Refs()
	assert.Equal(t, schemas.Jobs{
		"foo": newJob,
	}, refs[ref.Key()].LatestJobs)

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics()
	labels := map[string]string{
		"project":            ref.ProjectName,
		"topics":             ref.Topics,
		"ref":                ref.Name,
		"kind":               string(ref.Kind),
		"variables":          ref.LatestPipeline.Variables,
		"stage":              newJob.Stage,
		"job_name":           newJob.Name,
		"runner_description": ref.PullPipelineJobsRunnerDescriptionAggregationRegexp,
	}

	lastJobRunID := schemas.Metric{
		Kind:   schemas.MetricKindJobID,
		Labels: labels,
		Value:  2,
	}
	assert.Equal(t, lastJobRunID, metrics[lastJobRunID.Key()])

	timeSinceLastJobRun := schemas.Metric{
		Kind:   schemas.MetricKindJobTimestamp,
		Labels: labels,
		Value:  2,
	}
	assert.Equal(t, timeSinceLastJobRun, metrics[timeSinceLastJobRun.Key()])

	lastRunJobDuration := schemas.Metric{
		Kind:   schemas.MetricKindJobDurationSeconds,
		Labels: labels,
		Value:  newJob.DurationSeconds,
	}
	assert.Equal(t, lastRunJobDuration, metrics[lastRunJobDuration.Key()])

	jobRunCount := schemas.Metric{
		Kind:   schemas.MetricKindJobRunCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, jobRunCount, metrics[jobRunCount.Key()])

	artifactSize := schemas.Metric{
		Kind:   schemas.MetricKindJobArtifactSizeBytes,
		Labels: labels,
		Value:  float64(150),
	}
	assert.Equal(t, artifactSize, metrics[artifactSize.Key()])

	labels["status"] = newJob.Status
	status := schemas.Metric{
		Kind:   schemas.MetricKindJobStatus,
		Labels: labels,
		Value:  float64(1),
	}
	assert.Equal(t, status, metrics[status.Key()])
}
