package controller

import (
	"context"
	"strings"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// GetRefs ..
func (c *Controller) GetRefs(
	projectName string,
	filterRegexp string,
	maxAgeSeconds uint,
	fetchMergeRequestsPipelinesRefs bool,
	fetchMergeRequestsPipelinesRefsInitLimit int) (map[string]schemas.RefKind, error) {

	branches, err := c.Gitlab.GetProjectBranches(projectName, filterRegexp, maxAgeSeconds)
	if err != nil {
		return nil, err
	}

	tags, err := c.Gitlab.GetProjectTags(projectName, filterRegexp, maxAgeSeconds)
	if err != nil {
		return nil, err
	}

	mergeRequests := []string{}
	if fetchMergeRequestsPipelinesRefs {
		mergeRequests, err = c.Gitlab.GetProjectMergeRequestsPipelines(projectName, fetchMergeRequestsPipelinesRefsInitLimit, maxAgeSeconds)
		if err != nil {
			return nil, err
		}
	}

	foundRefs := map[string]schemas.RefKind{}
	for kind, refs := range map[schemas.RefKind][]string{
		schemas.RefKindBranch:       branches,
		schemas.RefKindTag:          tags,
		schemas.RefKindMergeRequest: mergeRequests,
	} {
		for _, ref := range refs {
			if _, ok := foundRefs[ref]; ok {
				log.Warn("found duplicate ref for project")
				continue
			}
			foundRefs[ref] = kind
		}
	}
	return foundRefs, nil
}

// PullRefsFromProject ..
func (c *Controller) PullRefsFromProject(ctx context.Context, p config.Project) error {
	gp, err := c.Gitlab.GetProject(p.Name)
	if err != nil {
		return err
	}

	refs, err := c.GetRefs(
		p.Name,
		p.Pull.Refs.Regexp,
		p.Pull.Refs.MaxAgeSeconds,
		p.Pull.Refs.From.MergeRequests.Enabled,
		int(p.Pull.Refs.From.MergeRequests.Depth),
	)
	if err != nil {
		return err
	}

	for ref, kind := range refs {
		ref := schemas.NewRef(
			kind,
			p.Name,
			ref,
			strings.Join(gp.TagList, ","),
			p.OutputSparseStatusMetrics,
			p.Pull.Pipeline.Jobs.Enabled,
			p.Pull.Pipeline.Jobs.FromChildPipelines.Enabled,
			p.Pull.Pipeline.Jobs.RunnerDescription.Enabled,
			p.Pull.Pipeline.Variables.Enabled,
			p.Pull.Pipeline.Variables.Regexp,
			p.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp,
		)

		refExists, err := c.Store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
				"ref-kind":     ref.Kind,
			}).Info("discovered new ref")

			if err = c.Store.SetRef(ref); err != nil {
				return err
			}

			c.ScheduleTask(ctx, TaskTypePullRefMetrics, ref)
		}
	}
	return nil
}

// PullRefsFromPipelines ..
func (c *Controller) PullRefsFromPipelines(ctx context.Context, p config.Project) error {
	log.WithFields(log.Fields{
		"init-operation": true,
		"project-name":   p.Name,
	}).Debug("fetching project")

	gp, err := c.Gitlab.GetProject(p.Name)
	if err != nil {
		return err
	}

	refs, err := c.Gitlab.GetRefsFromPipelines(p, strings.Join(gp.TagList, ","))
	if err != nil {
		return err
	}

	// Immediately trigger a pull of the ref
	for _, ref := range refs {
		refExists, err := c.Store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
				"ref-kind":     ref.Kind,
			}).Info("discovered new ref from pipelines")

			if err = c.Store.SetRef(ref); err != nil {
				return err
			}

			c.ScheduleTask(ctx, TaskTypePullRefMetrics, ref)
		}
	}
	return nil
}
