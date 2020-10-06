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

func TestPollProjectRefPipelineJobs(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()
	ConfigureGitlabClient(c)
	ConfigureStore()

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
	}

	assert.NoError(t, pollProjectRefPipelineJobs(pr))
	server.Close()
	assert.Error(t, pollProjectRefPipelineJobs(pr))
}

func TestPollProjectRefMostRecentJobs(t *testing.T) {
	mux, server, c := getMockedGitlabClient()
	defer server.Close()
	ConfigureGitlabClient(c)
	ConfigureStore()

	mux.HandleFunc("/api/v4/projects/1/jobs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1,"created_at":"2016-08-11T11:28:34.085Z"},{"id":2,"created_at":"2016-08-11T11:28:34.085Z"}]`)
		})

	pr := schemas.ProjectRef{
		ID:  1,
		Ref: "foo",
		Jobs: map[string]goGitlab.Job{
			"foo": goGitlab.Job{
				ID: 1,
			},
		},
	}

	// Test with FetchPipelineJobMetrics disabled
	assert.NoError(t, pollProjectRefMostRecentJobs(pr))

	// Enable FetchPipelineJobMetrics
	pr.FetchPipelineJobMetricsValue = pointy.Bool(true)
	assert.NoError(t, pollProjectRefMostRecentJobs(pr))
	server.Close()
	assert.Error(t, pollProjectRefMostRecentJobs(pr))
}

func TestProcessJobMetrics(t *testing.T) {
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
			Parameters: schemas.Parameters{
				OutputSparseStatusMetricsValue: pointy.Bool(true),
			},
		},
	}

	ConfigureStore()
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
		Kind:   schemas.MetricKindLastJobRunID,
		Labels: labels,
		Value:  2,
	}
	assert.Equal(t, lastJobRunID, metrics[lastJobRunID.Key()])

	timeSinceLastJobRun := schemas.Metric{
		Kind:   schemas.MetricKindTimeSinceLastJobRun,
		Labels: labels,
		Value:  time.Since(*newJob.CreatedAt).Round(time.Second).Seconds(),
	}
	assert.Equal(t, timeSinceLastJobRun, metrics[timeSinceLastJobRun.Key()])

	lastRunJobDuration := schemas.Metric{
		Kind:   schemas.MetricKindLastRunJobDuration,
		Labels: labels,
		Value:  newJob.Duration,
	}
	assert.Equal(t, lastRunJobDuration, metrics[lastRunJobDuration.Key()])

	timeSinceLastRun := schemas.Metric{
		Kind:   schemas.MetricKindTimeSinceLastRun,
		Labels: labels,
		Value:  time.Since(*newJob.CreatedAt).Round(time.Second).Seconds(),
	}
	assert.Equal(t, timeSinceLastRun, metrics[timeSinceLastRun.Key()])

	jobRunCount := schemas.Metric{
		Kind:   schemas.MetricKindJobRunCount,
		Labels: labels,
		Value:  float64(1),
	}
	assert.Equal(t, jobRunCount, metrics[jobRunCount.Key()])

	artifactSize := schemas.Metric{
		Kind:   schemas.MetricKindLastRunJobArtifactSize,
		Labels: labels,
		Value:  float64(150),
	}
	assert.Equal(t, artifactSize, metrics[artifactSize.Key()])

	labels["status"] = newJob.Status
	status := schemas.Metric{
		Kind:   schemas.MetricKindLastRunJobStatus,
		Labels: labels,
		Value:  float64(1),
	}
	assert.Equal(t, status, metrics[status.Key()])
}
