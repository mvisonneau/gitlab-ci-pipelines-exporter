package gitlab

import (
	"context"
	"regexp"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// GetProjectEnvironments ..
func (c *Client) GetProjectEnvironments(ctx context.Context, p schemas.Project) (
	envs schemas.Environments,
	err error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectEnvironments")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", p.Name))

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
		c.rateLimit(ctx)

		var (
			glenvs []*goGitlab.Environment
			resp   *goGitlab.Response
		)

		glenvs, resp, err = c.Environments.ListEnvironments(p.Name, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

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
func (c *Client) GetEnvironment(
	ctx context.Context,
	project string,
	environmentID int,
) (
	environment schemas.Environment,
	err error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetEnvironment")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", project))
	span.SetAttributes(attribute.Int("environment_id", environmentID))

	environment = schemas.Environment{
		ProjectName: project,
		ID:          environmentID,
	}

	c.rateLimit(ctx)

	var (
		e    *goGitlab.Environment
		resp *goGitlab.Response
	)

	e, resp, err = c.Environments.GetEnvironment(project, environmentID, goGitlab.WithContext(ctx))
	if err != nil || e == nil {
		return
	}

	c.requestsRemaining(resp)

	environment.Name = e.Name
	environment.ExternalURL = e.ExternalURL

	if e.State == "available" {
		environment.Available = true
	}

	if e.LastDeployment == nil {
		log.WithContext(ctx).
			WithFields(log.Fields{
				"project-name":     project,
				"environment-name": e.Name,
			}).
			Debug("no deployments found for the environment")

		return
	}

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

	return
}
