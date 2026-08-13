package gitlab

import (
	"context"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// GetProjectHooks ..
func (c *Client) GetProjectHooks(ctx context.Context, projectName string) (hooks []*goGitlab.ProjectHook, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProjectHooks")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", projectName))

	log.WithField("project_name", projectName).Trace("listing project hooks")

	c.rateLimit(ctx)

	hooks, resp, err := c.Projects.ListProjectHooks(
		projectName,
		&goGitlab.ListProjectHooksOptions{},
		goGitlab.WithContext(ctx),
	)
	if err != nil {
		return
	}

	c.requestsRemaining(resp)

	return hooks, nil
}

// AddProjectHook ..
func (c *Client) AddProjectHook(ctx context.Context, projectName string, options *goGitlab.AddProjectHookOptions) (hook *goGitlab.ProjectHook, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:AddProjectHook")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", projectName))

	log.WithField("project_name", projectName).Trace("adding project hook")

	c.rateLimit(ctx)

	hook, resp, err := c.Projects.AddProjectHook(
		projectName,
		options,
		goGitlab.WithContext(ctx),
	)
	if err != nil {
		return
	}

	c.requestsRemaining(resp)

	return hook, nil
}

// RemoveProjectHook ..
func (c *Client) RemoveProjectHook(ctx context.Context, projectName string, hookID int) (err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:RemoveProjectHook")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", projectName))
	span.SetAttributes(attribute.Int("hook_id", hookID))

	log.WithFields(log.Fields{
		"project_name": projectName,
		"hook_id":      hookID,
	}).Trace("removing project hook")

	c.rateLimit(ctx)

	resp, err := c.Projects.DeleteProjectHook(
		projectName,
		hookID,
		goGitlab.WithContext(ctx),
	)
	if err != nil {
		return
	}

	c.requestsRemaining(resp)

	return nil
}
