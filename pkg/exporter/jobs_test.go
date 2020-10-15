package exporter

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestPullProjectRefPipelineJobsMetrics(t *testing.T) {
	resetGlobalValues()

	mux, server := configureMockedGitlabClient()
	defer server.Close()
	configureStore()

	mux.HandleFunc("/api/v4/projects/1/pipelines/1/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "foo",
		MostRecentPipeline: &goGitlab.Pipeline{
			ID: 1,
		},
		Jobs: make(map[string]goGitlab.Job),
	}

	assert.NoError(t, pullProjectRefPipelineJobsMetrics(pr))
	server.Close()
	assert.Error(t, pullProjectRefPipelineJobsMetrics(pr))
}

func TestPullProjectRefMostRecentJobsMetrics(t *testing.T) {
	resetGlobalValues()

	mux, server := configureMockedGitlabClient()
	defer server.Close()
	configureStore()

	mux.HandleFunc("/api/v4/projects/1/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "foo",
		Jobs: map[string]goGitlab.Job{
			"foo": {
				ID: 1,
			},
		},
	}

	// Test with FetchPipelineJobMetrics disabled
	assert.NoError(t, pullProjectRefMostRecentJobsMetrics(pr))

	// Enable FetchPipelineJobMetrics
	pr.Pull.Pipeline.Jobs.EnabledValue = pointy.Bool(true)
	assert.NoError(t, pullProjectRefMostRecentJobsMetrics(pr))
	server.Close()
	assert.Error(t, pullProjectRefMostRecentJobsMetrics(pr))
}

func TestProcessJobMetrics(t *testing.T) {
	resetGlobalValues()
	configureStore()

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

	pr := schemas.ProjectRef{
		ID:                1,
		PathWithNamespace: "foo/bar",
		Topics:            "first,second",
		Kind:              schemas.ProjectRefKindBranch,
		Ref:               "foo",
		Jobs: map[string]goGitlab.Job{
			"foo": oldJob,
		},
		MostRecentPipeline: &goGitlab.Pipeline{
			ID: 1,
		},
		MostRecentPipelineVariables: "none",
		Project: schemas.Project{
			ProjectParameters: schemas.ProjectParameters{
				OutputSparseStatusMetricsValue: pointy.Bool(true),
			},
		},
	}

	store.SetProjectRef(pr)

	// If we run it against the same job, nothing should change in the store
	processJobMetrics(pr, oldJob)
	prs, _ := store.ProjectsRefs()
	assert.Equal(t, map[string]goGitlab.Job{
		"foo": oldJob,
	}, prs[pr.Key()].Jobs)

	// Update the project ref
	processJobMetrics(pr, newJob)
	prs, _ = store.ProjectsRefs()
	assert.Equal(t, map[string]goGitlab.Job{
		"foo": newJob,
	}, prs[pr.Key()].Jobs)

	// Check if all the metrics exist
	metrics, _ := store.Metrics()
	labels := map[string]string{
		"project":   pr.PathWithNamespace,
		"topics":    pr.Topics,
		"ref":       pr.Ref,
		"kind":      string(pr.Kind),
		"variables": pr.MostRecentPipelineVariables,
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
		Value:  float64(1),
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
