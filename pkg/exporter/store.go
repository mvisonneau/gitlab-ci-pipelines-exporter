package exporter

import (
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func garbageCollectProjects() error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	storedProjects, err := store.Projects()
	if err != nil {
		return err
	}

	// Loop through all configured projects
	for _, p := range config.Projects {
		delete(storedProjects, p.Key())
	}

	// Loop through what can be found from the wildcards
	for _, w := range config.Wildcards {
		foundProjects, err := gitlabClient.ListProjects(w)
		if err != nil {
			return err
		}

		for _, p := range foundProjects {
			delete(storedProjects, p.Key())
		}
	}

	log.WithFields(log.Fields{
		"projects-count": len(storedProjects),
	}).Info("found projects to garbage collect")

	for k, p := range storedProjects {
		if err = store.DelProject(k); err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"project-name": p.Name,
		}).Info("deleted project from the store")
	}

	return nil
}

func garbageCollectEnvironments() error {
	return nil
}

func garbageCollectRefs() error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	storedProjects, err := store.Projects()
	if err != nil {
		return err
	}

	storedRefs, err := store.Refs()
	if err != nil {
		return err
	}

	for k, ref := range storedRefs {
		p, projectExists := storedProjects[ref.Project.Key()]

		// If the project does not exist anymore, delete the project ref
		if !projectExists {
			if err = store.DelRef(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name": ref.PathWithNamespace,
				"project-ref":  ref.Ref,
				"reason":       "non-existent-project",
			}).Info("deleted project ref from the store")
			continue
		}

		// If the ref is not configured to be pulled anymore, delete the project ref
		re := regexp.MustCompile(p.ProjectParameters.Pull.Refs.Regexp())
		if !re.MatchString(ref.Ref) {
			if err = store.DelRef(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"project-name": ref.PathWithNamespace,
				"project-ref":  ref.Ref,
				"reason":       "ref-not-in-regexp",
			}).Info("deleted project ref from the store")
			continue
		}

		// Check if the latest configuration of the project in store matches the
		// ref one
		// TODO: Remove the storage of the project within the ref
		if ref.Project != p {
			ref.Project = p
			if err = store.SetRef(ref); err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"project-name": ref.PathWithNamespace,
				"project-ref":  ref.Ref,
				"reason":       "ref-not-in-regexp",
			}).Info("updated project ref, project definition was not in sync")
		}
	}

	return nil
}

func garbageCollectMetrics() error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	storedRefs, err := store.Refs()
	if err != nil {
		return err
	}

	storedMetrics, err := store.Metrics()
	if err != nil {
		return err
	}

	for k, m := range storedMetrics {
		// In order to save some memory space we chose to have to recompose
		// the Ref the metric belongs to
		metricLabelProject, metricLabelProjectExists := m.Labels["project"]
		metricLabelRef, metricLabelRefExists := m.Labels["ref"]

		if !metricLabelProjectExists || !metricLabelRefExists {
			if err = store.DelMetric(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"metric-kind":   m.Kind,
				"metric-labels": m.Labels,
				"reason":        "project-or-ref-label-undefined",
			}).Info("deleted metric from the store")
		}

		ref := schemas.Ref{
			PathWithNamespace: metricLabelProject,
			Ref:               metricLabelRef,
		}

		ref, refExists := storedRefs[ref.Key()]

		// If the project ref does not exist anymore, delete the metric
		if !refExists {
			if err = store.DelMetric(k); err != nil {
				return err
			}

			log.WithFields(log.Fields{
				"metric-kind":   m.Kind,
				"metric-labels": m.Labels,
				"reason":        "non-existent-project-ref",
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

			if !ref.Pull.Pipeline.Jobs.Enabled() {
				if err = store.DelMetric(k); err != nil {
					return err
				}

				log.WithFields(log.Fields{
					"metric-kind":   m.Kind,
					"metric-labels": m.Labels,
					"reason":        "jobs-metrics-disabled-on-project-ref",
				}).Info("deleted metric from the store")
				continue
			}

		default:
		}

		// Check if 'output sparse statuses metrics' has been enabled
		switch m.Kind {
		case schemas.MetricKindJobStatus,
			schemas.MetricKindStatus:

			if ref.OutputSparseStatusMetrics() && m.Value != 1 {
				if err = store.DelMetric(k); err != nil {
					return err
				}

				log.WithFields(log.Fields{
					"metric-kind":   m.Kind,
					"metric-labels": m.Labels,
					"reason":        "output-sparse-metrics-enabled-on-project-ref",
				}).Info("deleted metric from the store")
				continue
			}

		default:
		}
	}

	return nil
}

func metricLogFields(m schemas.Metric) log.Fields {
	return log.Fields{
		"metric-kind":   m.Kind,
		"metric-labels": m.Labels,
	}
}

func storeGetMetric(m *schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.GetMetric(m); err != nil {
		log.WithFields(
			metricLogFields(*m),
		).WithField(
			"error", err.Error(),
		).Errorf("reading metric from the store")
	}
}

func storeSetMetric(m schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.SetMetric(m); err != nil {
		log.WithFields(
			metricLogFields(m),
		).WithField(
			"error", err.Error(),
		).Errorf("writing metric in the store")
	}
}

func storeDelMetric(m schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.DelMetric(m.Key()); err != nil {
		log.WithFields(
			metricLogFields(m),
		).WithField(
			"error", err.Error(),
		).Errorf("deleting metric from the store")
	}
}
