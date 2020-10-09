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

			go schedulePullProjectRefsFromProject(context.Background(), p)
			if p.Pull.Refs.From.Pipelines.Enabled() {
				go schedulePullProjectRefsFromPipeline(context.Background(), p)
			}
		}
	}
	return nil
}
