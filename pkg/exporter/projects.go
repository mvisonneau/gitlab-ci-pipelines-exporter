package exporter

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func getProjectsFromWildcard(w schemas.Wildcard) error {
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

			if Config.OnInitFetchRefsFromPipelines {
				go pollingQueue.Add(getProjectRefsFromPipelinesTask.WithArgs(context.Background(), p))
			}

			go pollingQueue.Add(getRefsFromProjectTask.WithArgs(context.Background(), p))
		}
	}
	return nil
}
