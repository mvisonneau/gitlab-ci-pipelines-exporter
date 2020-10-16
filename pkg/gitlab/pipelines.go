package gitlab

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	goGitlab "github.com/xanzy/go-gitlab"
)

const (
	mergeRequestRefRegexp = `^refs/merge-requests`
)

// GetProjectRefPipeline ..
func (c *Client) GetProjectRefPipeline(pr schemas.ProjectRef, pipelineID int) (pipeline *goGitlab.Pipeline, err error) {
	c.rateLimit()
	pipeline, _, err = c.Pipelines.GetPipeline(pr.ID, pipelineID)
	if err != nil || pipeline == nil {
		return nil, fmt.Errorf("could not read content of pipeline %s - %s | %s", pr.PathWithNamespace, pr.Ref, err.Error())
	}
	return
}

// GetProjectPipelines ..
func (c *Client) GetProjectPipelines(projectID int, options *goGitlab.ListProjectPipelinesOptions) ([]*goGitlab.PipelineInfo, error) {
	fields := log.Fields{
		"project-id": projectID,
	}

	if options.Page == 0 {
		options.Page = 1
	}

	if options.PerPage == 0 {
		options.PerPage = 100
	}

	if options.Ref != nil {
		fields["project-ref"] = *options.Ref
	}

	if options.Scope != nil {
		fields["scope"] = *options.Scope
	}

	log.WithFields(fields).Debug("listing project pipelines")

	c.rateLimit()
	pipelines, _, err := c.Pipelines.ListProjectPipelines(projectID, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project ID %d: %s", projectID, err.Error())
	}
	return pipelines, nil
}

// GetProjectMergeRequestsPipelines ..
func (c *Client) GetProjectMergeRequestsPipelines(projectID int, fetchLimit int) ([]string, error) {
	var names []string

	options := &goGitlab.ListProjectPipelinesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re := regexp.MustCompile(mergeRequestRefRegexp)

	for {
		c.rateLimit()
		pipelines, resp, err := c.Pipelines.ListProjectPipelines(projectID, options)
		if err != nil {
			return nil, fmt.Errorf("error listing project pipelines for project ID %d: %s", projectID, err.Error())
		}

		for _, pipeline := range pipelines {
			if re.MatchString(pipeline.Ref) {
				names = append(names, pipeline.Ref)
				if len(names) >= fetchLimit {
					return names, nil
				}
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

// GetProjectRefPipelineVariablesAsConcatenatedString ..
func (c *Client) GetProjectRefPipelineVariablesAsConcatenatedString(pr schemas.ProjectRef) (string, error) {
	if pr.MostRecentPipeline == nil {
		log.WithFields(
			log.Fields{
				"project-id":  pr.ID,
				"project-ref": pr.Ref,
			},
		).Debug("most recent pipeline not defined, exiting..")
		return "", nil
	}

	log.WithFields(
		log.Fields{
			"project-name": pr.PathWithNamespace,
			"project-id":   pr.ID,
			"pipeline-id":  pr.MostRecentPipeline.ID,
		},
	).Debug("fetching pipeline variables")

	variablesFilter, err := regexp.Compile(pr.Pull.Pipeline.Variables.Regexp())
	if err != nil {
		return "", fmt.Errorf("the provided filter regex for pipeline variables is invalid '(%s)': %v", pr.Pull.Pipeline.Variables.Regexp(), err)
	}

	c.rateLimit()
	variables, _, err := c.Pipelines.GetPipelineVariables(pr.ID, pr.MostRecentPipeline.ID)
	if err != nil {
		return "", fmt.Errorf("could not fetch pipeline variables for %d: %s", pr.MostRecentPipeline.ID, err.Error())
	}

	var keptVariables []string
	if len(variables) > 0 {
		for _, v := range variables {
			if variablesFilter.MatchString(v.Key) {
				keptVariables = append(keptVariables, strings.Join([]string{v.Key, v.Value}, ":"))
			}
		}
	}

	return strings.Join(keptVariables, ","), nil
}

// GetProjectRefsFromPipelines ..
func (c *Client) GetProjectRefsFromPipelines(p schemas.Project, gp *goGitlab.Project) (schemas.ProjectsRefs, error) {
	re, err := regexp.Compile(p.Pull.Refs.Regexp())
	if err != nil {
		return nil, err
	}

	options := &goGitlab.ListProjectPipelinesOptions{
		ListOptions: goGitlab.ListOptions{
			Page: 1,
			// TODO: Get a proper loop to split this query up
			PerPage: p.Pull.Refs.From.Pipelines.Depth(),
		},
		Scope: pointy.String("branches"),
	}

	branchPipelines, err := c.GetProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, err
	}

	options.Scope = pointy.String("tags")
	tagsPipelines, err := c.GetProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, err
	}

	projectRefs := make(schemas.ProjectsRefs)
	for kind, pipelines := range map[schemas.ProjectRefKind][]*gitlab.PipelineInfo{
		schemas.ProjectRefKindBranch: branchPipelines,
		schemas.ProjectRefKindTag:    tagsPipelines,
	} {
		for _, pipeline := range pipelines {
			if re.MatchString(pipeline.Ref) {
				pr := schemas.NewProjectRef(p, gp, pipeline.Ref, kind)
				if _, ok := projectRefs[pr.Key()]; !ok {
					log.WithFields(
						log.Fields{
							"project-id":       gp.ID,
							"project-name":     gp.PathWithNamespace,
							"project-ref":      pipeline.Ref,
							"project-ref-kind": kind,
						},
					).Info("found project ref")
					projectRefs[pr.Key()] = pr
				}
			}
		}
	}

	return projectRefs, nil
}
