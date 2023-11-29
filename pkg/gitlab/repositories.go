package gitlab

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.openly.dev/pointy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// GetCommitCountBetweenRefs ..
func (c *Client) GetCommitCountBetweenRefs(ctx context.Context, project, from, to string) (int, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetCommitCountBetweenRefs")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", project))
	span.SetAttributes(attribute.String("from_ref", from))
	span.SetAttributes(attribute.String("to_ref", to))

	log.WithFields(log.Fields{
		"project-name": project,
		"from-ref":     from,
		"to-ref":       to,
	}).Debug("comparing refs")

	c.rateLimit(ctx)

	cmp, resp, err := c.Repositories.Compare(project, &goGitlab.CompareOptions{
		From:     &from,
		To:       &to,
		Straight: pointy.Bool(true),
	}, goGitlab.WithContext(ctx))
	if err != nil {
		return 0, err
	}

	c.requestsRemaining(resp)

	if cmp == nil {
		return 0, fmt.Errorf("could not compare refs successfully")
	}

	return len(cmp.Commits), nil
}
