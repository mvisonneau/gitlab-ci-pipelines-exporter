package gitlab

import (
	"regexp"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectEnvironmentIDs ..
func (c *Client) GetProjectEnvironmentIDs(project, envRegexp string) ([]int, error) {
	envIDs := []int{}

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
			return envIDs, err
		}

		for _, env := range envs {
			if re.MatchString(env.Name) {
				envIDs = append(envIDs, env.ID)
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return envIDs, nil
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
