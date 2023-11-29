package controller

import (
	"context"

	"dario.cat/mergo"
	log "github.com/sirupsen/logrus"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// GetRefs ..
func (c *Controller) GetRefs(ctx context.Context, p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	var pulledRefs schemas.Refs

	refs = make(schemas.Refs)

	if p.Pull.Refs.Branches.Enabled {
		// If one of these parameter is set, we will need to fetch the branches from the
		// pipelines API instead of the branches one
		if !p.Pull.Refs.Branches.ExcludeDeleted ||
			p.Pull.Refs.Branches.MostRecent > 0 ||
			p.Pull.Refs.Branches.MaxAgeSeconds > 0 {
			if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(ctx, p, schemas.RefKindBranch); err != nil {
				return
			}
		} else {
			if pulledRefs, err = c.Gitlab.GetProjectBranches(ctx, p); err != nil {
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
			if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(ctx, p, schemas.RefKindTag); err != nil {
				return
			}
		} else {
			if pulledRefs, err = c.Gitlab.GetProjectTags(ctx, p); err != nil {
				return
			}
		}

		if err = mergo.Merge(&refs, pulledRefs); err != nil {
			return
		}
	}

	if p.Pull.Refs.MergeRequests.Enabled {
		if pulledRefs, err = c.Gitlab.GetRefsFromPipelines(
			ctx,
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
	refs, err := c.GetRefs(ctx, p)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		refExists, err := c.Store.RefExists(ctx, ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
				"ref-kind":     ref.Kind,
			}).Info("discovered new ref")

			if err = c.Store.SetRef(ctx, ref); err != nil {
				return err
			}

			c.ScheduleTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), ref)
		}
	}

	return nil
}
