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

// GetProjectBranches ..
func (c *Client) GetProjectBranches(ctx context.Context, p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectBranches")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", p.Name))

	refs = make(schemas.Refs)

	options := &goGitlab.ListBranchesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var re *regexp.Regexp

	if re, err = regexp.Compile(p.Pull.Refs.Branches.Regexp); err != nil {
		return
	}

	for {
		c.rateLimit(ctx)

		var (
			branches []*goGitlab.Branch
			resp     *goGitlab.Response
		)

		branches, resp, err = c.Branches.ListBranches(p.Name, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		for _, branch := range branches {
			if re.MatchString(branch.Name) {
				ref := schemas.NewRef(p, schemas.RefKindBranch, branch.Name)
				refs[ref.Key()] = ref
			}
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}

		options.Page = resp.NextPage
	}

	return
}

// GetBranchLatestCommit ..
func (c *Client) GetBranchLatestCommit(ctx context.Context, project, branch string) (string, float64, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetBranchLatestCommit")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", project))
	span.SetAttributes(attribute.String("branch_name", branch))

	log.WithFields(log.Fields{
		"project-name": project,
		"branch":       branch,
	}).Debug("reading project branch")

	c.rateLimit(ctx)

	b, resp, err := c.Branches.GetBranch(project, branch, goGitlab.WithContext(ctx))
	if err != nil {
		return "", 0, err
	}

	c.requestsRemaining(resp)

	return b.Commit.ShortID, float64(b.Commit.CommittedDate.Unix()), nil
}
