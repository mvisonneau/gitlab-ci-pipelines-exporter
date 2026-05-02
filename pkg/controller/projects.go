package controller

import (
	"context"
	"sync"
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

	// Process project discovery in parallel using a bounded worker pool.
	// maxWorkers limits concurrent store operations to avoid overwhelming the store.
	const maxWorkers = 20
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, p := range foundProjects {
		p := p // capture loop variable for goroutine
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			projectExists, err := c.Store.ProjectExists(ctx, p.Key())
			if err != nil || projectExists {
				if err != nil {
					log.WithContext(ctx).WithError(err).Error()
				}
				return
			}

			log.WithFields(log.Fields{
				"wildcard-search":                  w.Search,
				"wildcard-owner-kind":              w.Owner.Kind,
				"wildcard-owner-name":              w.Owner.Name,
				"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
				"wildcard-archived":                w.Archived,
				"project-name":                     p.Name,
			}).Info("discovered new project")

			if err := c.Store.SetProject(ctx, p); err != nil {
				log.WithContext(ctx).WithError(err).Error()
				return
			}

			// Schedule ref + environment discovery immediately for new projects
			c.ScheduleTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()), p)
			c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()), p)
		}()
	}

	wg.Wait()
	return nil
}