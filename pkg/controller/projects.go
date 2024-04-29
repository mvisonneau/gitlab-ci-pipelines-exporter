package controller

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// PullProject ..
func (c *Controller) PullProject(ctx context.Context, name string, pull config.ProjectPull) error {
	gp, err := c.Gitlab.GetProject(ctx, name)
	if err != nil {
		return err
	}

	p := schemas.NewProject(gp.PathWithNamespace)
	p.Pull = pull

	projectExists, err := c.Store.ProjectExists(ctx, p.Key())
	if err != nil {
		return err
	}

	if !projectExists {
		log.WithFields(log.Fields{
			"project-name": p.Name,
		}).Info("discovered new project")

		if err := c.Store.SetProject(ctx, p); err != nil {
			log.WithContext(ctx).
				WithError(err).
				Error()
		}

		c.ScheduleTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()), p)
		c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()), p)
	}

	return nil
}

// PullProjectsFromWildcard ..
func (c *Controller) PullProjectsFromWildcard(ctx context.Context, w config.Wildcard) error {
	foundProjects, err := c.Gitlab.ListProjects(ctx, w)
	if err != nil {
		return err
	}

	for _, p := range foundProjects {
		projectExists, err := c.Store.ProjectExists(ctx, p.Key())
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

			if err := c.Store.SetProject(ctx, p); err != nil {
				log.WithContext(ctx).
					WithError(err).
					Error()
			}

			c.ScheduleTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()), p)
			c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()), p)
		}
	}

	return nil
}
