package controller

import (
	"context"

	"github.com/imdario/mergo"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// GetRefs ..
func (c *Controller) GetRefs(p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	refs = make(schemas.Refs)
	var pulledRefs schemas.Refs

	if p.Pull.Refs.Branches.Enabled {
		// If one of these parameter is set, we will need to fetch the branches from the
		// pipelines API instead of the branches one
		if !p.Pull.Refs.Branches.ExcludeDeleted ||
			p.Pull.Refs.Branches.MostRecent > 0 ||
			p.Pull.Refs.Branches.MaxAgeSeconds > 0 {

			if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(p, schemas.RefKindBranch); err != nil {
				return
			}
		} else {
			if pulledRefs, err = c.Gitlab.GetProjectBranches(p); err != nil {
				return
			}
		}

		if err = mergo.Merge(&refs, pulledRefs); err != nil {
			return
		}
	}

	if p.Pull.Refs.Tags.Enabled {
		// If one of these parameter is set, we will need to fetch the tags from the
		// pipelines API instead of the tags one
		if !p.Pull.Refs.Tags.ExcludeDeleted ||
			p.Pull.Refs.Tags.MostRecent > 0 ||
			p.Pull.Refs.Tags.MaxAgeSeconds > 0 {

			if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(p, schemas.RefKindTag); err != nil {
				return
			}
		} else {
			if pulledRefs, err = c.Gitlab.GetProjectTags(p); err != nil {
				return
			}
		}

		if err = mergo.Merge(&refs, pulledRefs); err != nil {
			return
		}
	}

	if p.Pull.Refs.MergeRequests.Enabled {
		if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(
			p,
			schemas.RefKindMergeRequest,
		); err != nil {
			return
		}

		if err = mergo.Merge(&refs, pulledRefs); err != nil {
			return
		}
	}

	return
}

// PullRefsFromProject ..
func (c *Controller) PullRefsFromProject(ctx context.Context, p schemas.Project) error {
	refs, err := c.GetRefs(p)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		refExists, err := c.Store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-name": ref.Project.Name,
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
func (c *Controller) PullRefsFromPipelines(ctx context.Context, p schemas.Project) error {
	log.WithFields(log.Fields{
		"init-operation": true,
		"project-name":   p.Name,
	}).Debug("fetching project")

	refs := make(schemas.Refs)
	if p.Pull.Refs.Branches.Enabled {
		branches, err := c.Gitlab.GetRefsFromPipelines(p, schemas.RefKindBranch)
		if err != nil {
			return err
		}
		for _, ref := range branches {
			refs[ref.Key()] = ref
		}
	}

	if p.Pull.Refs.Tags.Enabled {
		tags, err := c.Gitlab.GetRefsFromPipelines(p, schemas.RefKindTag)
		if err != nil {
			return err
		}
		for _, ref := range tags {
			refs[ref.Key()] = ref
		}
	}

	if p.Pull.Refs.MergeRequests.Enabled {
		mrs, err := c.Gitlab.GetRefsFromPipelines(p, schemas.RefKindMergeRequest)
		if err != nil {
			return err
		}
		for _, ref := range mrs {
			refs[ref.Key()] = ref
		}
	}

	// Immediately trigger a pull of the ref
	for _, ref := range refs {
		refExists, err := c.Store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-name": ref.Project.Name,
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
