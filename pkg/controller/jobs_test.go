package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestPullRefPipelineJobsMetrics(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z","started_at":"2016-08-11T11:28:56.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z","started_at":"2016-08-11T11:28:58.085Z"}]`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Jobs.FromChildPipelines.Enabled = false

	ref := schemas.NewRef(p, schemas.RefKindBranch, "bar")
	ref.LatestPipeline.ID = 1

	// TODO: assert the results?
	assert.NoError(t, c.PullRefPipelineJobsMetrics(ctx, ref))
	srv.Close()
	assert.Error(t, c.PullRefPipelineJobsMetrics(ctx, ref))
}

func TestPullRefMostRecentJobsMetrics(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	ref := schemas.Ref{
		Project: schemas.NewProject("foo"),
		Name:    "bar",
		LatestJobs: schemas.Jobs{
			"bar": {
				ID: 1,
			},
		},
	}

	// Test with FetchPipelineJobMetrics disabled
	assert.NoError(t, c.PullRefMostRecentJobsMetrics(ctx, ref))

	// Enable FetchPipelineJobMetrics
	ref.Project.Pull.Pipeline.Jobs.Enabled = true
	assert.NoError(t, c.PullRefMostRecentJobsMetrics(ctx, ref))
	srv.Close()
	assert.Error(t, c.PullRefMostRecentJobsMetrics(ctx, ref))
}

func TestProcessJobMetrics(t *testing.T) {
	ctx, c, _, srv := newTestController(config.Config{})
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
		TagList:         "",
		ArtifactSize:    150,
		Runner: schemas.Runner{
			Description: "foo-123-bar",
		},
	}

	p := schemas.NewProject("foo")
	p.Topics = "first,second"
	p.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp = `foo-(.*)-bar`

	ref := schemas.NewRef(p, schemas.RefKindBranch, "foo")
	ref.LatestPipeline.ID = 1
	ref.LatestPipeline.Variables = "none"
	ref.LatestJobs = schemas.Jobs{
		"foo": oldJob,
	}

	c.Store.SetRef(ctx, ref)

	// If we run it against the same job, nothing should change in the store
	c.ProcessJobMetrics(ctx, ref, oldJob)
	refs, _ := c.Store.Refs(ctx)
	assert.Equal(t, schemas.Jobs{
		"foo": oldJob,
	}, refs[ref.Key()].LatestJobs)

	// Update the ref
	c.ProcessJobMetrics(ctx, ref, newJob)
	refs, _ = c.Store.Refs(ctx)
	assert.Equal(t, schemas.Jobs{
		"foo": newJob,
	}, refs[ref.Key()].LatestJobs)

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics(ctx)
	labels := map[string]string{
		"project":            ref.Project.Name,
		"topics":             ref.Project.Topics,
		"ref":                ref.Name,
		"kind":               string(ref.Kind),
		"variables":          ref.LatestPipeline.Variables,
		"source":             ref.LatestPipeline.Source,
		"stage":              newJob.Stage,
		"tag_list":           newJob.TagList,
		"failure_reason":     newJob.FailureReason,
		"job_name":           newJob.Name,
		"runner_description": ref.Project.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp,
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
