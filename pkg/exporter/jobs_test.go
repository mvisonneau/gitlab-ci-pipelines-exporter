package exporter

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
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
		MostRecentPipeline: &goGitlab.Pipeline{
			ID: 1,
		},
		Jobs: make(map[string]goGitlab.Job),
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
		Jobs: map[string]goGitlab.Job{
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

	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)
	oldJob := goGitlab.Job{
		ID:        1,
		Name:      "foo",
		CreatedAt: &oneDayAgo,
	}

	newJob := goGitlab.Job{
		ID:        2,
		Name:      "foo",
		CreatedAt: &now,
		Duration:  15,
		Status:    "failed",
		Stage:     "🚀",
		Artifacts: []struct {
			FileType   string "json:\"file_type\""
			Filename   string "json:\"filename\""
			Size       int    "json:\"size\""
			FileFormat string "json:\"file_format\""
		}{
			{
				Size: 100,
			},
			{
				Size: 50,
			},
		},
	}

	ref := schemas.Ref{
		ProjectName: "foo/bar",
		Topics:      "first,second",
		Kind:        schemas.RefKindBranch,
		Name:        "foo",
		Jobs: map[string]goGitlab.Job{
			"foo": oldJob,
		},
		MostRecentPipeline: &goGitlab.Pipeline{
			ID: 1,
		},
		MostRecentPipelineVariables: "none",
		OutputSparseStatusMetrics:   true,
	}

	store.SetRef(ref)

	// If we run it against the same job, nothing should change in the store
	processJobMetrics(ref, oldJob)
	refs, _ := store.Refs()
	assert.Equal(t, map[string]goGitlab.Job{
		"foo": oldJob,
	}, refs[ref.Key()].Jobs)

	// Update the project ref
	processJobMetrics(ref, newJob)
	refs, _ = store.Refs()
	assert.Equal(t, map[string]goGitlab.Job{
		"foo": newJob,
	}, refs[ref.Key()].Jobs)

	// Check if all the metrics exist
	metrics, _ := store.Metrics()
	labels := map[string]string{
		"project":   ref.ProjectName,
		"topics":    ref.Topics,
		"ref":       ref.Name,
		"kind":      string(ref.Kind),
		"variables": ref.MostRecentPipelineVariables,
		"stage":     newJob.Stage,
		"job_name":  newJob.Name,
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
		Value:  float64(now.Unix()),
	}
	assert.Equal(t, timeSinceLastJobRun, metrics[timeSinceLastJobRun.Key()])

	lastRunJobDuration := schemas.Metric{
		Kind:   schemas.MetricKindJobDurationSeconds,
		Labels: labels,
		Value:  newJob.Duration,
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
