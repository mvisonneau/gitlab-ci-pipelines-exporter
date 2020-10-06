package exporter

import (
	"context"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func getProjectRefs(
	projectID int,
	refsRegexp string,
	fetchMergeRequestsPipelinesRefs bool,
	fetchMergeRequestsPipelinesRefsInitLimit int) (map[string]schemas.ProjectRefKind, error) {

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

	foundRefs := map[string]schemas.ProjectRefKind{}
	for kind, refs := range map[schemas.ProjectRefKind][]string{
		schemas.ProjectRefKindBranch:       branches,
		schemas.ProjectRefKindTag:          tags,
		schemas.ProjectRefKindMergeRequest: mergeRequests,
	} {
		for _, ref := range refs {
			if _, ok := foundRefs[ref]; ok {
				log.Warn("found duplicate ref for project")
			}
			foundRefs[ref] = kind
		}
	}
	return foundRefs, nil
}

func getRefsFromProject(p schemas.Project) {
	gp, err := gitlabClient.GetProject(p.Name)
	if err != nil {
		log.WithFields(
			log.Fields{
				"project-name": p.Name,
				"error":        err.Error(),
			},
		).Errorf("getting project by name")
		return
	}

	refs, err := getProjectRefs(
		gp.ID,
		p.RefsRegexp(),
		p.FetchMergeRequestsPipelinesRefs(),
		p.FetchMergeRequestsPipelinesRefsLimit(),
	)

	if err != nil {
		log.WithFields(
			log.Fields{
				"project-id":                  gp.ID,
				"project-path-with-namespace": gp.PathWithNamespace,
				"error":                       err.Error(),
			},
		).Error("getting project refs")
		return
	}

	for ref, kind := range refs {
		pr := schemas.NewProjectRef(p, gp, ref, kind)
		projectRefExists, err := store.ProjectRefExists(pr.Key())
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if !projectRefExists {
			log.WithFields(
				log.Fields{
					"project-id":                  gp.ID,
					"project-path-with-namespace": gp.PathWithNamespace,
					"project-ref":                 ref,
					"project-ref-kind":            kind,
				},
			).Info("discovered new project ref")

			store.SetProjectRef(pr)
			go pollingQueue.Add(pollProjectRefMostRecentPipelineTask.WithArgs(context.Background(), pr))

			if pr.FetchPipelineJobMetrics() {
				go pollingQueue.Add(pollProjectRefMostRecentJobsTask.WithArgs(context.Background(), pr))
			}
		}
	}
}

func getProjectRefsFromPipelines(p schemas.Project) {
	log.WithFields(
		log.Fields{
			"init-operation": true,
			"project-name":   p.Name,
		},
	).Debug("fetching project")

	gp, err := gitlabClient.GetProject(p.Name)
	if err != nil {
		log.WithFields(
			log.Fields{
				"init-operation": true,
				"project-name":   p.Name,
				"error":          err.Error(),
			},
		).Errorf("getting GitLab project by name")
		return
	}

	projectRefs, err := gitlabClient.GetProjectRefsFromPipelines(p, gp, Config.OnInitFetchRefsFromPipelinesDepthLimit)
	if err != nil {
		log.WithFields(
			log.Fields{
				"init-operation": true,
				"project-name":   p.Name,
			},
		).Errorf("unable to fetch refs from project pipelines: %s", err.Error())
	}

	// Immediately trigger a poll of the ref
	for _, pr := range projectRefs {
		projectRefExists, err := store.ProjectRefExists(pr.Key())
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if !projectRefExists {
			log.WithFields(
				log.Fields{
					"project-id":                  gp.ID,
					"project-path-with-namespace": gp.PathWithNamespace,
					"project-ref":                 pr.Ref,
					"project-ref-kind":            pr.Kind,
				},
			).Info("discovered new project ref from pipelines")

			store.SetProjectRef(pr)
			go pollingQueue.Add(pollProjectRefMostRecentPipelineTask.WithArgs(context.Background(), pr))

			if pr.FetchPipelineJobMetrics() {
				go pollingQueue.Add(pollProjectRefMostRecentJobsTask.WithArgs(context.Background(), pr))
			}
		}
	}
}
