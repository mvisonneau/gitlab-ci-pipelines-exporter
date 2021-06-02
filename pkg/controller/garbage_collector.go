package controller

import (
	"context"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// GarbageCollectProjects ..
func (c *Controller) GarbageCollectProjects(_ context.Context) error {
	log.Info("starting 'projects' garbage collection")
	defer log.Info("ending 'projects' garbage collection")

	storedProjects, err := c.Store.Projects()
	if err != nil {
		return err
	}

	// Loop through all configured projects
	for _, p := range c.Projects {
		delete(storedProjects, p.Key())
	}

	// Loop through what can be found from the wildcards
	for _, w := range c.Wildcards {
		foundProjects, err := c.Gitlab.ListProjects(w)
		if err != nil {
			return err
		}

		for _, p := range foundProjects {
			delete(storedProjects, p.Key())
		}
	}

	log.WithFields(log.Fields{
		"projects-count": len(storedProjects),
	}).Debug("found projects to garbage collect")

	for k, p := range storedProjects {
		if err = c.Store.DelProject(k); err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"project-name": p.Name,
		}).Info("deleted project from the store")
	}

	return nil
}

// GarbageCollectEnvironments ..
func (c *Controller) GarbageCollectEnvironments(_ context.Context) error {
	log.Info("starting 'environments' garbage collection")
	defer log.Info("ending 'environments' garbage collection")

	storedEnvironments, err := c.Store.Environments()
	if err != nil {
		return err
	}

	envProjects := make(map[string]string)
	for k, env := range storedEnvironments {
		p := config.Project{
			Name: env.ProjectName,
		}

		projectExists, err := c.Store.ProjectExists(p.Key())
		if err != nil {
			return err
		}

		// If the project does not exist anymore, delete the environment
		if !projectExists {
			if err = c.Store.DelEnvironment(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-name": env.Name,
				"reason":           "non-existent-project",
			}).Info("deleted environment from the store")
			continue
		}

		if err = c.Store.GetProject(&p); err != nil {
			return err
		}

		// Store the project information to be able to refresh its environments
		// from the API later on
		envProjects[p.Name] = p.Pull.Environments.Regexp

		// If the environment is not configured to be pulled anymore, delete it
		re := regexp.MustCompile(p.Pull.Environments.Regexp)
		if !re.MatchString(env.Name) {
			if err = c.Store.DelEnvironment(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-name": env.Name,
				"reason":           "environment-not-in-regexp",
			}).Info("deleted environment from the store")
			continue
		}

		// Check if the latest configuration of the project in store matches the environment one
		if env.OutputSparseStatusMetrics != p.OutputSparseStatusMetrics {
			env.OutputSparseStatusMetrics = p.OutputSparseStatusMetrics

			if err = c.Store.SetEnvironment(env); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-name": env.Name,
			}).Info("updated ref, associated project configuration was not in sync")
		}
	}

	// Refresh the environments from the API
	existingEnvs := make(map[schemas.EnvironmentKey]struct{})
	for projectName, envRegexp := range envProjects {
		envs, err := c.Gitlab.GetProjectEnvironments(projectName, envRegexp)
		if err != nil {
			return err
		}

		for _, envName := range envs {
			existingEnvs[schemas.Environment{
				ProjectName: projectName,
				Name:        envName,
			}.Key()] = struct{}{}
		}
	}

	storedEnvironments, err = c.Store.Environments()
	if err != nil {
		return err
	}

	for k, env := range storedEnvironments {
		if _, exists := existingEnvs[k]; !exists {
			if err = c.Store.DelEnvironment(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-name": env.Name,
				"reason":           "non-existent-environment",
			}).Info("deleted environment from the store")
		}
	}

	return nil
}

// GarbageCollectRefs ..
func (c *Controller) GarbageCollectRefs(_ context.Context) error {
	log.Info("starting 'refs' garbage collection")
	defer log.Info("ending 'refs' garbage collection")

	storedRefs, err := c.Store.Refs()
	if err != nil {
		return err
	}

	refProjects := make(map[string]config.ProjectPullRefs)
	for k, ref := range storedRefs {
		p := config.Project{Name: ref.ProjectName}
		projectExists, err := c.Store.ProjectExists(p.Key())
		if err != nil {
			return err
		}

		// If the project does not exist anymore, delete the ref
		if !projectExists {
			if err = c.Store.DelRef(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
				"reason":       "non-existent-project",
			}).Info("deleted ref from the store")
			continue
		}

		if err = c.Store.GetProject(&p); err != nil {
			return err
		}

		// Store the project information to be able to refresh all refs
		// from the API later on
		refProjects[p.Name] = p.Pull.Refs

		// If the ref is not configured to be pulled anymore, delete the ref
		re := regexp.MustCompile(p.Pull.Refs.Regexp)
		if !re.MatchString(ref.Name) {
			if err = c.Store.DelRef(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
				"reason":       "ref-not-in-regexp",
			}).Info("deleted ref from the store")
			continue
		}

		// Check if the latest configuration of the project in store matches the ref one
		if ref.OutputSparseStatusMetrics != p.OutputSparseStatusMetrics ||
			ref.PullPipelineJobsEnabled != p.Pull.Pipeline.Jobs.Enabled ||
			ref.PullPipelineVariablesEnabled != p.Pull.Pipeline.Variables.Enabled ||
			ref.PullPipelineVariablesRegexp != p.Pull.Pipeline.Variables.Regexp {
			ref.OutputSparseStatusMetrics = p.OutputSparseStatusMetrics
			ref.PullPipelineJobsEnabled = p.Pull.Pipeline.Jobs.Enabled
			ref.PullPipelineVariablesEnabled = p.Pull.Pipeline.Variables.Enabled
			ref.PullPipelineVariablesRegexp = p.Pull.Pipeline.Variables.Regexp
			if err = c.Store.SetRef(ref); err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
			}).Info("updated ref, associated project configuration was not in sync")
		}
	}

	// Refresh the refs from the API
	existingRefs := make(map[schemas.RefKey]struct{})
	for projectName, projectPullRefs := range refProjects {
		branches, err := c.Gitlab.GetProjectBranches(projectName, projectPullRefs.Regexp, projectPullRefs.MaxAgeSeconds)
		if err != nil {
			return err
		}

		for _, branch := range branches {
			existingRefs[schemas.Ref{
				Kind:        schemas.RefKindBranch,
				ProjectName: projectName,
				Name:        branch,
			}.Key()] = struct{}{}
		}

		tags, err := c.Gitlab.GetProjectTags(projectName, projectPullRefs.Regexp, projectPullRefs.MaxAgeSeconds)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			existingRefs[schemas.Ref{
				Kind:        schemas.RefKindTag,
				ProjectName: projectName,
				Name:        tag,
			}.Key()] = struct{}{}
		}

		if projectPullRefs.From.MergeRequests.Enabled {
			mergeRequests, err := c.Gitlab.GetProjectMergeRequestsPipelines(projectName, int(projectPullRefs.From.MergeRequests.Depth), projectPullRefs.MaxAgeSeconds)
			if err != nil {
				return err
			}

			for _, mr := range mergeRequests {
				existingRefs[schemas.Ref{
					Kind:        schemas.RefKindMergeRequest,
					ProjectName: projectName,
					Name:        mr,
				}.Key()] = struct{}{}
			}
		}
	}

	storedRefs, err = c.Store.Refs()
	if err != nil {
		return err
	}

	for k, ref := range storedRefs {
		if _, exists := existingRefs[k]; !exists {
			if err = c.Store.DelRef(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref-name":     ref.Name,
				"reason":       "non-existent-ref",
			}).Info("deleted ref from the store")
		}
	}

	return nil
}

// GarbageCollectMetrics ..
func (c *Controller) GarbageCollectMetrics(_ context.Context) error {
	log.Info("starting 'metrics' garbage collection")
	defer log.Info("ending 'metrics' garbage collection")

	storedEnvironments, err := c.Store.Environments()
	if err != nil {
		return err
	}

	storedRefs, err := c.Store.Refs()
	if err != nil {
		return err
	}

	storedMetrics, err := c.Store.Metrics()
	if err != nil {
		return err
	}

	for k, m := range storedMetrics {
		// In order to save some memory space we chose to have to recompose
		// the Ref the metric belongs to
		metricLabelProject, metricLabelProjectExists := m.Labels["project"]
		metricLabelRef, metricLabelRefExists := m.Labels["ref"]
		metricLabelEnvironment, metricLabelEnvironmentExists := m.Labels["environment"]

		if !metricLabelProjectExists || (!metricLabelRefExists && !metricLabelEnvironmentExists) {
			if err = c.Store.DelMetric(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"metric-kind":   m.Kind,
				"metric-labels": m.Labels,
				"reason":        "project-or-ref-and-environment-label-undefined",
			}).Info("deleted metric from the store")
		}

		if metricLabelRefExists && !metricLabelEnvironmentExists {
			refKey := schemas.Ref{
				Kind:        schemas.RefKind(m.Labels["kind"]),
				ProjectName: metricLabelProject,
				Name:        metricLabelRef,
			}.Key()

			ref, refExists := storedRefs[refKey]

			// If the ref does not exist anymore, delete the metric
			if !refExists {
				if err = c.Store.DelMetric(k); err != nil {
					return err
				}

				log.WithFields(log.Fields{
					"metric-kind":   m.Kind,
					"metric-labels": m.Labels,
					"reason":        "non-existent-ref",
				}).Info("deleted metric from the store")
				continue
			}

			// Check if the pulling of jobs related metrics has been disabled
			switch m.Kind {
			case schemas.MetricKindJobArtifactSizeBytes,
				schemas.MetricKindJobDurationSeconds,
				schemas.MetricKindJobID,
				schemas.MetricKindJobRunCount,
				schemas.MetricKindJobStatus,
				schemas.MetricKindJobTimestamp:

				if !ref.PullPipelineJobsEnabled {
					if err = c.Store.DelMetric(k); err != nil {
						return err
					}

					log.WithFields(log.Fields{
						"metric-kind":   m.Kind,
						"metric-labels": m.Labels,
						"reason":        "jobs-metrics-disabled-on-ref",
					}).Info("deleted metric from the store")
					continue
				}

			default:
			}

			// Check if 'output sparse statuses metrics' has been enabled
			switch m.Kind {
			case schemas.MetricKindJobStatus,
				schemas.MetricKindStatus:

				if ref.OutputSparseStatusMetrics && m.Value != 1 {
					if err = c.Store.DelMetric(k); err != nil {
						return err
					}

					log.WithFields(log.Fields{
						"metric-kind":   m.Kind,
						"metric-labels": m.Labels,
						"reason":        "output-sparse-metrics-enabled-on-ref",
					}).Info("deleted metric from the store")
					continue
				}

			default:
			}
		}

		if metricLabelEnvironmentExists {
			envKey := schemas.Environment{
				ProjectName: metricLabelProject,
				Name:        metricLabelEnvironment,
			}.Key()

			env, envExists := storedEnvironments[envKey]

			// If the ref does not exist anymore, delete the metric
			if !envExists {
				if err = c.Store.DelMetric(k); err != nil {
					return err
				}

				log.WithFields(log.Fields{
					"metric-kind":   m.Kind,
					"metric-labels": m.Labels,
					"reason":        "non-existent-environment",
				}).Info("deleted metric from the store")
				continue
			}

			// Check if 'output sparse statuses metrics' has been enabled
			switch m.Kind {
			case schemas.MetricKindEnvironmentDeploymentStatus:
				if env.OutputSparseStatusMetrics && m.Value != 1 {
					if err = c.Store.DelMetric(k); err != nil {
						return err
					}

					log.WithFields(log.Fields{
						"metric-kind":   m.Kind,
						"metric-labels": m.Labels,
						"reason":        "output-sparse-metrics-enabled-on-environment",
					}).Info("deleted metric from the store")
					continue
				}
			}
		}
	}

	return nil
}
