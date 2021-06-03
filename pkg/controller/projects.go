package controller

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	log "github.com/sirupsen/logrus"
)

// PullProjectsFromWildcard ..
func (c *Controller) PullProjectsFromWildcard(ctx context.Context, w config.Wildcard) error {
	foundProjects, err := c.Gitlab.ListProjects(w)
	if err != nil {
		return err
	}

	for _, p := range foundProjects {
		projectExists, err := c.Store.ProjectExists(p.Key())
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

			if err := c.Store.SetProject(p); err != nil {
				log.Errorf(err.Error())
			}

			c.ScheduleTask(ctx, TaskTypePullRefsFromProject, p)
			c.ScheduleTask(ctx, TaskTypePullEnvironmentsFromProject, p)
		}
	}

	return nil
}
