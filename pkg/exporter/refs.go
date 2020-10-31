package exporter

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func getRefs(
	projectID int,
	refsRegexp string,
	fetchMergeRequestsPipelinesRefs bool,
	fetchMergeRequestsPipelinesRefsInitLimit int) (map[string]schemas.RefKind, error) {

	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	branches, err := gitlabClient.GetProjectBranches(projectID, refsRegexp)
	if err != nil {
		return nil, err
	}

	tags, err := gitlabClient.GetProjectTags(projectID, refsRegexp)
	if err != nil {
		return nil, err
	}

	mergeRequests := []string{}
	if fetchMergeRequestsPipelinesRefs {
		mergeRequests, err = gitlabClient.GetProjectMergeRequestsPipelines(projectID, fetchMergeRequestsPipelinesRefsInitLimit)
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

func pullRefsFromProject(p schemas.Project) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	gp, err := gitlabClient.GetProject(p.Name)
	if err != nil {
		return err
	}

	refs, err := getRefs(
		gp.ID,
		p.Pull.Refs.Regexp(),
		p.Pull.Refs.From.MergeRequests.Enabled(),
		p.Pull.Refs.From.MergeRequests.Depth(),
	)

	if err != nil {
		return err
	}

	for ref, kind := range refs {
		ref := schemas.NewRef(p, gp, ref, kind)
		projectRefExists, err := store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !projectRefExists {
			log.WithFields(log.Fields{
				"project-id":       gp.ID,
				"project-name":     gp.PathWithNamespace,
				"project-ref":      ref.Ref,
				"project-ref-kind": kind,
			}).Info("discovered new project ref")

			if err = store.SetRef(ref); err != nil {
				return err
			}

			go schedulePullRefMetrics(context.Background(), ref)
		}
	}
	return nil
}

func pullRefsFromPipelines(p schemas.Project) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	log.WithFields(log.Fields{
		"init-operation": true,
		"project-name":   p.Name,
	}).Debug("fetching project")

	gp, err := gitlabClient.GetProject(p.Name)
	if err != nil {
		return err
	}

	refs, err := gitlabClient.GetRefsFromPipelines(p, gp)
	if err != nil {
		return err
	}

	// Immediately trigger a pull of the ref
	for _, ref := range refs {
		refExists, err := store.RefExists(ref.Key())
		if err != nil {
			return err
		}

		if !refExists {
			log.WithFields(log.Fields{
				"project-id":       gp.ID,
				"project-name":     gp.PathWithNamespace,
				"project-ref":      ref.Ref,
				"project-ref-kind": ref.Kind,
			}).Info("discovered new project ref from pipelines")

			if err = store.SetRef(ref); err != nil {
				return err
			}

			go schedulePullRefMetrics(context.Background(), ref)
		}
	}
	return nil
}
