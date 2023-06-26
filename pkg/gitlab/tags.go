package gitlab

import (
	"context"
	"regexp"

	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// GetProjectTags ..
func (c *Client) GetProjectTags(ctx context.Context, p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectTags")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", p.Name))

	refs = make(schemas.Refs)

	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var re *regexp.Regexp

	if re, err = regexp.Compile(p.Pull.Refs.Tags.Regexp); err != nil {
		return
	}

	for {
		c.rateLimit(ctx)

		var (
			tags []*goGitlab.Tag
			resp *goGitlab.Response
		)

		tags, resp, err = c.Tags.ListTags(p.Name, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				ref := schemas.NewRef(p, schemas.RefKindTag, tag.Name)
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

// GetProjectMostRecentTagCommit ..
func (c *Client) GetProjectMostRecentTagCommit(ctx context.Context, projectName, filterRegexp string) (string, float64, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectTags")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", projectName))
	span.SetAttributes(attribute.String("regexp", filterRegexp))

	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re, err := regexp.Compile(filterRegexp)
	if err != nil {
		return "", 0, err
	}

	for {
		c.rateLimit(ctx)

		tags, resp, err := c.Tags.ListTags(projectName, options, goGitlab.WithContext(ctx))
		if err != nil {
			return "", 0, err
		}

		c.requestsRemaining(resp)

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				return tag.Commit.ShortID, float64(tag.Commit.CommittedDate.Unix()), nil
			}
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}

		options.Page = resp.NextPage
	}

	return "", 0, nil
}
