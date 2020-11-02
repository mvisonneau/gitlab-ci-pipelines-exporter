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

// GetRefPipeline ..
func (c *Client) GetRefPipeline(ref schemas.Ref, pipelineID int) (pipeline *goGitlab.Pipeline, err error) {
	c.rateLimit()
	pipeline, _, err = c.Pipelines.GetPipeline(ref.ProjectName, pipelineID)
	if err != nil || pipeline == nil {
		return nil, fmt.Errorf("could not read content of pipeline %s - %s | %s", ref.ProjectName, ref.Name, err.Error())
	}
	return
}

// GetProjectPipelines ..
func (c *Client) GetProjectPipelines(projectName string, options *goGitlab.ListProjectPipelinesOptions) ([]*goGitlab.PipelineInfo, error) {
	fields := log.Fields{
		"project-name": projectName,
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
	pipelines, _, err := c.Pipelines.ListProjectPipelines(projectName, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project %s: %s", projectName, err.Error())
	}
	return pipelines, nil
}

// GetProjectMergeRequestsPipelines ..
func (c *Client) GetProjectMergeRequestsPipelines(projectName string, fetchLimit int) ([]string, error) {
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
		pipelines, resp, err := c.Pipelines.ListProjectPipelines(projectName, options)
		if err != nil {
			return nil, fmt.Errorf("error listing project pipelines for project %s: %s", projectName, err.Error())
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

// GetRefPipelineVariablesAsConcatenatedString ..
func (c *Client) GetRefPipelineVariablesAsConcatenatedString(ref schemas.Ref) (string, error) {
	if ref.MostRecentPipeline == nil {
		log.WithFields(
			log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting..")
		return "", nil
	}

	log.WithFields(
		log.Fields{
			"project-name": ref.ProjectName,
			"ref":          ref.Name,
			"pipeline-id":  ref.MostRecentPipeline.ID,
		},
	).Debug("fetching pipeline variables")

	variablesFilter, err := regexp.Compile(ref.PullPipelineVariablesRegexp)
	if err != nil {
		return "", fmt.Errorf("the provided filter regex for pipeline variables is invalid '(%s)': %v", ref.PullPipelineVariablesRegexp, err)
	}

	c.rateLimit()
	variables, _, err := c.Pipelines.GetPipelineVariables(ref.ProjectName, ref.MostRecentPipeline.ID)
	if err != nil {
		return "", fmt.Errorf("could not fetch pipeline variables for %d: %s", ref.MostRecentPipeline.ID, err.Error())
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

// GetRefsFromPipelines ..
func (c *Client) GetRefsFromPipelines(p schemas.Project, topics string) (schemas.Refs, error) {
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

	branchPipelines, err := c.GetProjectPipelines(p.Name, options)
	if err != nil {
		return nil, err
	}

	options.Scope = pointy.String("tags")
	tagsPipelines, err := c.GetProjectPipelines(p.Name, options)
	if err != nil {
		return nil, err
	}

	refs := make(schemas.Refs)
	for kind, pipelines := range map[schemas.RefKind][]*gitlab.PipelineInfo{
		schemas.RefKindBranch: branchPipelines,
		schemas.RefKindTag:    tagsPipelines,
	} {
		for _, pipeline := range pipelines {
			if re.MatchString(pipeline.Ref) {
				ref := schemas.NewRef(
					kind,
					p.Name,
					pipeline.Ref,
					topics,
					p.OutputSparseStatusMetrics(),
					p.Pull.Pipeline.Jobs.Enabled(),
					p.Pull.Pipeline.Variables.Enabled(),
					p.Pull.Pipeline.Variables.Regexp(),
				)

				if _, ok := refs[ref.Key()]; !ok {
					log.WithFields(
						log.Fields{
							"project-name": p.Name,
							"ref":          pipeline.Ref,
							"ref-kind":     kind,
						},
					).Info("found ref")
					refs[ref.Key()] = ref
				}
			}
		}
	}

	return refs, nil
}
