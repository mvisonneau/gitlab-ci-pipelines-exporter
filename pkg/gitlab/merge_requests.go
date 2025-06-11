package gitlab

import (
	"context"
	"regexp"
	"strconv"

	goGitlab "gitlab.com/gitlab-org/api/client-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/utils"
)

// GetProjectOpenMergeRequests ..
func (c *Client) GetProjectOpenMergeRequests(ctx context.Context, p schemas.Project) (
	mrs schemas.Refs,
	err error,
) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectOpenMergeRequests")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", p.Name))

	mrs = make(schemas.Refs)

	options := &goGitlab.ListProjectMergeRequestsOptions{
		State: utils.Ptr("opened"),
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var re *regexp.Regexp

	if re, err = regexp.Compile(p.Pull.Refs.MergeRequests.Regexp); err != nil {
		return
	}

	for {
		c.rateLimit(ctx)

		var (
			mrsList []*goGitlab.BasicMergeRequest
			resp    *goGitlab.Response
		)

		mrsList, resp, err = c.MergeRequests.ListProjectMergeRequests(p.Name, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		for _, mr := range mrsList {
			if re.MatchString(mr.Title) {
				ref := schemas.NewRef(p, schemas.RefKindMergeRequest, strconv.Itoa(mr.IID))
				mrs[ref.Key()] = ref
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return
}
