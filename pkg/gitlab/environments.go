package gitlab

import (
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectEnvironments ..
func (c *Client) GetProjectEnvironments(p schemas.Project) (
	envs schemas.Environments,
	err error,
) {
	envs = make(schemas.Environments)

	options := &goGitlab.ListEnvironmentsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	if p.Pull.Environments.ExcludeStopped {
		options.States = goGitlab.String("available")
	}

	re, err := regexp.Compile(p.Pull.Environments.Regexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		var glenvs []*goGitlab.Environment
		var resp *goGitlab.Response
		glenvs, resp, err = c.Environments.ListEnvironments(p.Name, options)
		if err != nil {
			return
		}

		for _, glenv := range glenvs {
			if re.MatchString(glenv.Name) {
				env := schemas.Environment{
					ProjectName:               p.Name,
					ID:                        glenv.ID,
					Name:                      glenv.Name,
					OutputSparseStatusMetrics: p.OutputSparseStatusMetrics,
				}

				if glenv.State == "available" {
					env.Available = true
				}

				envs[env.Key()] = env
			}
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}
		options.Page = resp.NextPage
	}

	return
}

// GetEnvironment ..
func (c *Client) GetEnvironment(project string, environmentID int) (schemas.Environment, error) {
	environment := schemas.Environment{
		ProjectName: project,
		ID:          environmentID,
	}

	c.rateLimit()
	e, _, err := c.Environments.GetEnvironment(project, environmentID, nil)
	if err != nil || e == nil {
		return environment, err
	}

	environment.Name = e.Name
	environment.ExternalURL = e.ExternalURL

	if e.State == "available" {
		environment.Available = true
	}

	if e.LastDeployment != nil {
		if e.LastDeployment.Deployable.Tag {
			environment.LatestDeployment.RefKind = schemas.RefKindTag
		} else {
			environment.LatestDeployment.RefKind = schemas.RefKindBranch
		}

		environment.LatestDeployment.RefName = e.LastDeployment.Ref
		environment.LatestDeployment.JobID = e.LastDeployment.Deployable.ID
		environment.LatestDeployment.DurationSeconds = e.LastDeployment.Deployable.Duration
		environment.LatestDeployment.Status = e.LastDeployment.Deployable.Status

		if e.LastDeployment.Deployable.User != nil {
			environment.LatestDeployment.Username = e.LastDeployment.Deployable.User.Username
		}

		if e.LastDeployment.Deployable.Commit != nil {
			environment.LatestDeployment.CommitShortID = e.LastDeployment.Deployable.Commit.ShortID
		}

		if e.LastDeployment.CreatedAt != nil {
			environment.LatestDeployment.Timestamp = float64(e.LastDeployment.CreatedAt.Unix())
		}
	} else {
		log.WithFields(log.Fields{
			"project-name":     project,
			"environment-name": e.Name,
		}).Warn("no deployments found for the environment")
	}

	return environment, nil
}
