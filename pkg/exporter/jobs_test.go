package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestPullRefPipelineJobsMetrics(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

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

	assert.NoError(t, pullRefPipelineJobsMetrics(ref))
	server.Close()
	assert.Error(t, pullRefPipelineJobsMetrics(ref))
}

func TestPullRefMostRecentJobsMetrics(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

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
	assert.NoError(t, pullRefMostRecentJobsMetrics(ref))

	// Enable FetchPipelineJobMetrics
	ref.PullPipelineJobsEnabled = true
	assert.NoError(t, pullRefMostRecentJobsMetrics(ref))
	server.Close()
	assert.Error(t, pullRefMostRecentJobsMetrics(ref))
}

func TestProcessJobMetrics(t *testing.T) {
	resetGlobalValues()

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
			Description: "xxx",
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
		OutputSparseStatusMetrics: true,
	}

	store.SetRef(ref)

	// If we run it against the same job, nothing should change in the store
	processJobMetrics(ref, oldJob)
	refs, _ := store.Refs()
	assert.Equal(t, schemas.Jobs{
		"foo": oldJob,
	}, refs[ref.Key()].LatestJobs)

	// Update the ref
	processJobMetrics(ref, newJob)
	refs, _ = store.Refs()
	assert.Equal(t, schemas.Jobs{
		"foo": newJob,
	}, refs[ref.Key()].LatestJobs)

	// Check if all the metrics exist
	metrics, _ := store.Metrics()
	labels := map[string]string{
		"project":            ref.ProjectName,
		"topics":             ref.Topics,
		"ref":                ref.Name,
		"kind":               string(ref.Kind),
		"variables":          ref.LatestPipeline.Variables,
		"stage":              newJob.Stage,
		"job_name":           newJob.Name,
		"runner_description": newJob.Runner.Description,
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
