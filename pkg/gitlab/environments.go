package gitlab

import (
	"regexp"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectEnvironments ..
func (c *Client) GetProjectEnvironments(project, envRegexp string) (map[int]string, error) {
	environments := map[int]string{}

	options := &goGitlab.ListEnvironmentsOptions{
		Page:    1,
		PerPage: 100,
	}

	re, err := regexp.Compile(envRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		envs, resp, err := c.Environments.ListEnvironments(project, options)
		if err != nil {
			return environments, err
		}

		for _, env := range envs {
			if re.MatchString(env.Name) {
				environments[env.ID] = env.Name
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return environments, nil
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

		environment.LatestDeployment.ID = e.LastDeployment.ID
		environment.LatestDeployment.RefName = e.LastDeployment.Ref
		environment.LatestDeployment.AuthorEmail = e.LastDeployment.Deployable.User.PublicEmail
		environment.LatestDeployment.CommitShortID = e.LastDeployment.Deployable.Commit.ShortID
		environment.LatestDeployment.CreatedAt = *e.LastDeployment.CreatedAt
		environment.LatestDeployment.Duration = time.Duration(int(e.LastDeployment.Deployable.Duration)) * time.Second
		environment.LatestDeployment.Status = e.LastDeployment.Deployable.Status
	} else {
		log.WithFields(log.Fields{
			"project-name":     project,
			"environment-name": e.Name,
		}).Warn("no deployments found for the environment")
	}

	return environment, nil
}
