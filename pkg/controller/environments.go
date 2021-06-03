package controller

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// PullEnvironmentsFromProject ..
func (c *Controller) PullEnvironmentsFromProject(ctx context.Context, p schemas.Project) (err error) {
	var envs schemas.Environments
	envs, err = c.Gitlab.GetProjectEnvironments(p)
	if err != nil {
		return
	}

	for k := range envs {
		var exists bool
		exists, err = c.Store.EnvironmentExists(k)
		if err != nil {
			return
		}

		if !exists {
			env := envs[k]
			if err = c.UpdateEnvironment(&env); err != nil {
				return
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-id":   env.ID,
				"environment-name": env.Name,
			}).Info("discovered new environment")

			c.ScheduleTask(ctx, TaskTypePullEnvironmentMetrics, env)
		}
	}
	return
}

// UpdateEnvironment ..
func (c *Controller) UpdateEnvironment(env *schemas.Environment) error {
	pulledEnv, err := c.Gitlab.GetEnvironment(env.ProjectName, env.ID)
	if err != nil {
		return err
	}

	env.Available = pulledEnv.Available
	env.ExternalURL = pulledEnv.ExternalURL
	env.LatestDeployment = pulledEnv.LatestDeployment

	return c.Store.SetEnvironment(*env)
}

// PullEnvironmentMetrics ..
func (c *Controller) PullEnvironmentMetrics(env schemas.Environment) (err error) {
	// At scale, the scheduled environment may be behind the actual state being stored
	// to avoid issues, we refresh it from the store before manipulating it
	if err := c.Store.GetEnvironment(&env); err != nil {
		return err
	}

	// Save the existing deployment ID before we updated environment from the API
	deploymentJobID := env.LatestDeployment.JobID
	if err = c.UpdateEnvironment(&env); err != nil {
		return
	}

	infoLabels := env.InformationLabelsValues()
	var commitDate float64
	switch env.LatestDeployment.RefKind {
	case schemas.RefKindBranch:
		infoLabels["latest_commit_short_id"], commitDate, err = c.Gitlab.GetBranchLatestCommit(env.ProjectName, env.LatestDeployment.RefName)
	case schemas.RefKindTag:
		// TODO: Review how to manage this in a nicier fashion
		infoLabels["latest_commit_short_id"], commitDate, err = c.Gitlab.GetProjectMostRecentTagCommit(env.ProjectName, ".*")
	default:
		infoLabels["latest_commit_short_id"] = env.LatestDeployment.CommitShortID
		commitDate = env.LatestDeployment.Timestamp
	}

	if err != nil {
		return err
	}

	var (
		envBehindDurationSeconds float64
		envBehindCommitCount     float64
	)

	behindCommitsCountMetric := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindCommitsCount,
		Labels: env.DefaultLabelsValues(),
	}

	// To reduce the amount of compare requests being made, we check if the labels are unchanged since
	// the latest emission of the information metric
	if infoLabels["latest_commit_short_id"] != infoLabels["current_commit_short_id"] {
		infoMetric := schemas.Metric{
			Kind:   schemas.MetricKindEnvironmentInformation,
			Labels: env.DefaultLabelsValues(),
		}

		var commitCount int
		if err = c.Store.GetMetric(&infoMetric); err != nil {
			return err
		}

		if infoMetric.Labels["latest_commit_short_id"] != infoLabels["latest_commit_short_id"] ||
			infoMetric.Labels["current_commit_short_id"] != infoLabels["current_commit_short_id"] {
			commitCount, err = c.Gitlab.GetCommitCountBetweenRefs(env.ProjectName, infoLabels["current_commit_short_id"], infoLabels["latest_commit_short_id"])
			if err != nil {
				return err
			}
			envBehindCommitCount = float64(commitCount)
		} else {
			// TODO: Find a more efficient way
			if err = c.Store.GetMetric(&behindCommitsCountMetric); err != nil {
				return err
			}
			envBehindCommitCount = behindCommitsCountMetric.Value
		}
	}

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindCommitsCount,
		Labels: env.DefaultLabelsValues(),
		Value:  envBehindCommitCount,
	})

	if commitDate-env.LatestDeployment.Timestamp > 0 {
		envBehindDurationSeconds = commitDate - env.LatestDeployment.Timestamp
	}

	envDeploymentCount := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentCount,
		Labels: env.DefaultLabelsValues(),
	}

	storeGetMetric(c.Store, &envDeploymentCount)
	if env.LatestDeployment.JobID > deploymentJobID {
		envDeploymentCount.Value++
	}
	storeSetMetric(c.Store, envDeploymentCount)

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindDurationSeconds,
		Labels: env.DefaultLabelsValues(),
		Value:  envBehindDurationSeconds,
	})

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentDurationSeconds,
		Labels: env.DefaultLabelsValues(),
		Value:  env.LatestDeployment.DurationSeconds,
	})

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentJobID,
		Labels: env.DefaultLabelsValues(),
		Value:  float64(env.LatestDeployment.JobID),
	})

	emitStatusMetric(
		c.Store,
		schemas.MetricKindEnvironmentDeploymentStatus,
		env.DefaultLabelsValues(),
		statusesList[:],
		env.LatestDeployment.Status,
		env.OutputSparseStatusMetrics,
	)

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentTimestamp,
		Labels: env.DefaultLabelsValues(),
		Value:  env.LatestDeployment.Timestamp,
	})

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentInformation,
		Labels: infoLabels,
		Value:  1,
	})

	return nil
}
