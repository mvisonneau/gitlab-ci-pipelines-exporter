package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestPullEnvironmentsFromProject(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/environments"),
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"name":"main"},{"id":1337,"name":"prod"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/environments/1337",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
{
	"id": 1,
	"name": "prod",
	"external_url": "https://foo.example.com",
	"state": "available",
	"last_deployment": {
		"ref": "bar",
		"created_at": "2019-03-25T18:55:13.252Z",
		"deployable": {
			"id": 2,
			"status": "success",
			"tag": false,
			"duration": 21623.13423,
			"user": {
				"username": "alice"
			},
			"commit": {
				"short_id": "416d8ea1"
			}
		}
	}
}`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Environments.Regexp = "^prod"
	assert.NoError(t, c.PullEnvironmentsFromProject(ctx, p))

	storedEnvironments, _ := c.Store.Environments(ctx)
	expectedEnvironments := schemas.Environments{
		"54146361": schemas.Environment{
			ProjectName: "foo",
			ID:          1337,
			Name:        "prod",
			ExternalURL: "https://foo.example.com",
			Available:   true,
			LatestDeployment: schemas.Deployment{
				JobID:           2,
				RefKind:         schemas.RefKindBranch,
				RefName:         "bar",
				Username:        "alice",
				Timestamp:       1553540113,
				DurationSeconds: 21623.13423,
				CommitShortID:   "416d8ea1",
				Status:          "success",
			},
			OutputSparseStatusMetrics: true,
		},
	}
	assert.Equal(t, expectedEnvironments, storedEnvironments)
}

func TestPullEnvironmentMetricsSucceed(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/environments/1",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
{
	"id": 1,
	"name": "prod",
	"external_url": "https://foo.example.com",
	"state": "available",
	"last_deployment": {
		"ref": "bar",
		"created_at": "2019-03-25T18:55:13.252Z",
		"deployable": {
			"id": 2,
			"status": "success",
			"tag": false,
			"duration": 21623.13423,
			"user": {
				"public_email": "foo@bar.com"
			},
			"commit": {
				"short_id": "416d8ea1"
			}
		}
	}
}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/repository/branches/bar",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
{
	"commit": {
		"short_id": "416d8ea1",
		"committed_date": "2019-03-25T18:55:13.252Z"
	}
}`)
		})

	env := schemas.Environment{
		ProjectName: "foo",
		Name:        "prod",
		ID:          1,
	}

	// Metrics pull shall succeed
	assert.NoError(t, c.PullEnvironmentMetrics(ctx, env))

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics(ctx)
	labels := map[string]string{
		"project":     "foo",
		"environment": "prod",
	}

	environmentBehindCommitsCount := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindCommitsCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, environmentBehindCommitsCount, metrics[environmentBehindCommitsCount.Key()])

	environmentBehindCommitsDurationSeconds := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindDurationSeconds,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, environmentBehindCommitsDurationSeconds, metrics[environmentBehindCommitsDurationSeconds.Key()])

	environmentDeploymentDurationSeconds := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentDurationSeconds,
		Labels: labels,
		Value:  21623.13423,
	}
	assert.Equal(t, environmentDeploymentDurationSeconds, metrics[environmentDeploymentDurationSeconds.Key()])

	labels["status"] = "success"
	status := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])
}
