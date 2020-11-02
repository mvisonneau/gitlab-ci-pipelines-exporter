package exporter

import (
	"context"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func pullEnvironmentsFromProject(p schemas.Project) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	envs, err := gitlabClient.GetProjectEnvironments(p.Name, p.Pull.Environments.NameRegexp())
	if err != nil {
		return err
	}

	for envID, envName := range envs {
		env := schemas.Environment{
			ProjectName: p.Name,
			Name:        envName,
			ID:          envID,

			TagsRegexp:                p.Pull.Environments.TagsRegexp(),
			OutputSparseStatusMetrics: p.OutputSparseStatusMetrics(),
		}

		envExists, err := store.EnvironmentExists(env.Key())
		if err != nil {
			return err
		}

		if !envExists {
			if err = updateEnvironment(&env); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-id":   env.ID,
				"environment-name": env.Name,
			}).Info("discovered new environment")

			go schedulePullEnvironmentMetrics(context.Background(), env)
		}
	}
	return nil
}

func updateEnvironment(env *schemas.Environment) error {
	pulledEnv, err := gitlabClient.GetEnvironment(env.ProjectName, env.ID)
	if err != nil {
		return err
	}

	env.Available = pulledEnv.Available
	env.ExternalURL = pulledEnv.ExternalURL
	env.LatestDeployment = pulledEnv.LatestDeployment

	return store.SetEnvironment(*env)
}

func pullEnvironmentMetrics(env schemas.Environment) (err error) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	// Save the existing deployment ID before we updated environment from the API
	deploymentID := env.LatestDeployment.ID
	if err = updateEnvironment(&env); err != nil {
		return
	}

	infoLabels := env.InformationLabelsValues()
	var commitDate time.Time
	switch env.LatestDeployment.RefKind {
	case schemas.RefKindBranch:
		infoLabels["latest_commit_short_id"], commitDate, err = gitlabClient.GetBranchLatestCommit(env.ProjectName, env.LatestDeployment.RefName)
	case schemas.RefKindTag:
		infoLabels["latest_commit_short_id"], commitDate, err = gitlabClient.GetBranchLatestCommit(env.ProjectName, env.LatestDeployment.RefName)
	default:
		infoLabels["latest_commit_short_id"] = env.LatestDeployment.CommitShortID
		commitDate = env.LatestDeployment.CreatedAt
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
		if err = store.GetMetric(&infoMetric); err != nil {
			return err
		}

		if infoMetric.Labels["latest_commit_short_id"] != infoLabels["latest_commit_short_id"] ||
			infoMetric.Labels["current_commit_short_id"] != infoLabels["current_commit_short_id"] {
			commitCount, err = gitlabClient.GetCommitCountBetweenRefs(env.ProjectName, infoLabels["current_commit_short_id"], infoLabels["latest_commit_short_id"])
			if err != nil {
				return err
			}
			envBehindCommitCount = float64(commitCount)
		} else {
			// TODO: Find a more efficient way
			if err = store.GetMetric(&behindCommitsCountMetric); err != nil {
				return err
			}
			envBehindCommitCount = behindCommitsCountMetric.Value
		}
	}

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindCommitsCount,
		Labels: env.DefaultLabelsValues(),
		Value:  envBehindCommitCount,
	})

	if commitDate.Sub(env.LatestDeployment.CreatedAt).Seconds() > 0 {
		envBehindDurationSeconds = commitDate.Sub(env.LatestDeployment.CreatedAt).Seconds()
	}

	envDeploymentCount := schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentCount,
		Labels: env.DefaultLabelsValues(),
	}

	storeGetMetric(&envDeploymentCount)
	if env.LatestDeployment.ID > deploymentID {
		envDeploymentCount.Value++
	}
	storeSetMetric(envDeploymentCount)

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentBehindDurationSeconds,
		Labels: env.DefaultLabelsValues(),
		Value:  envBehindDurationSeconds,
	})

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentDurationSeconds,
		Labels: env.DefaultLabelsValues(),
		Value:  env.LatestDeployment.Duration.Seconds(),
	})

	emitStatusMetric(
		schemas.MetricKindEnvironmentDeploymentStatus,
		env.DefaultLabelsValues(),
		statusesList[:],
		env.LatestDeployment.Status,
		env.OutputSparseStatusMetrics,
	)

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentDeploymentTimestamp,
		Labels: env.DefaultLabelsValues(),
		Value:  float64(env.LatestDeployment.CreatedAt.Unix()),
	})

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindEnvironmentInformation,
		Labels: infoLabels,
		Value:  1,
	})

	return nil
}
