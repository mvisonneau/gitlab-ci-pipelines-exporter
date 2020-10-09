package exporter

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func pullProjectsFromWildcard(w schemas.Wildcard) error {
	foundProjects, err := gitlabClient.ListProjects(w)
	if err != nil {
		return err
	}

	for _, p := range foundProjects {
		projectExists, err := store.ProjectExists(p.Key())
		if err != nil {
			return err
		}

		if !projectExists {
			log.WithFields(log.Fields{
				"wildcard-search":                  w.Search,
				"wildcard-owner-kind":              w.Owner.Kind,
				"wildcard-owner-name":              w.Owner.Name,
				"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
				"wildcard-archived":                w.Archived,
				"project-name":                     p.Name,
			}).Info("discovered new project")

			if err := store.SetProject(p); err != nil {
				log.Errorf(err.Error())
			}

			if p.Pull.Refs.From.Pipelines.Enabled() {
				if err = pullingQueue.Add(pullProjectRefsFromPipelinesTask.WithArgs(context.Background(), p)); err != nil {
					log.WithFields(log.Fields{
						"project-name": p.Name,
						"error":        err.Error(),
					}).Error("scheduling 'project refs from pipelines' pull")
				}
			}

			if err = pullingQueue.Add(pullProjectRefsFromProjectTask.WithArgs(context.Background(), p)); err != nil {
				log.WithFields(log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				}).Error("scheduling 'project refs from project' pull")
			}
		}
	}
	return nil
}
