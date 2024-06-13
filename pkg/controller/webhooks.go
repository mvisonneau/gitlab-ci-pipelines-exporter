package controller

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func (c *Controller) processPipelineEvent(ctx context.Context, e goGitlab.PipelineEvent) {
	var (
		refKind schemas.RefKind
		refName = e.ObjectAttributes.Ref
	)

	// TODO: Perhaps it would be nice to match upon the regexp to validate
	// that it is actually a merge request ref
	if e.MergeRequest.IID != 0 {
		refKind = schemas.RefKindMergeRequest
		refName = strconv.Itoa(e.MergeRequest.IID)
	} else if e.ObjectAttributes.Tag {
		refKind = schemas.RefKindTag
	} else {
		refKind = schemas.RefKindBranch
	}

	c.triggerRefMetricsPull(ctx, schemas.NewRef(
		schemas.NewProject(e.Project.PathWithNamespace),
		refKind,
		refName,
	))
}

func (c *Controller) processJobEvent(ctx context.Context, e goGitlab.JobEvent) {
	var (
		refKind schemas.RefKind
		refName = e.Ref
	)

	if e.Tag {
		refKind = schemas.RefKindTag
	} else {
		refKind = schemas.RefKindBranch
	}

	project, _, err := c.Gitlab.Projects.GetProject(e.ProjectID, nil)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error("reading project from GitLab")

		return
	}

	c.triggerRefMetricsPull(ctx, schemas.NewRef(
		schemas.NewProject(project.PathWithNamespace),
		refKind,
		refName,
	))
}

func (c *Controller) processPushEvent(ctx context.Context, e goGitlab.PushEvent) {
	if e.CheckoutSHA == "" {
		var (
			refKind = schemas.RefKindBranch
			refName string
		)

		// branch refs in push events have "refs/heads/" prefix
		if branch, found := strings.CutPrefix(e.Ref, "refs/heads/"); found {
			refName = branch
		} else {
			log.WithContext(ctx).
				WithFields(log.Fields{
					"project-name": e.Project.Name,
					"ref":          e.Ref,
				}).
				Error("extracting branch name from ref")

			return
		}

		_ = deleteRef(ctx, c.Store, schemas.NewRef(
			schemas.NewProject(e.Project.PathWithNamespace),
			refKind,
			refName,
		), "received branch deletion push event from webhook")
	}
}

func (c *Controller) processTagEvent(ctx context.Context, e goGitlab.TagEvent) {
	if e.CheckoutSHA == "" {
		var (
			refKind = schemas.RefKindTag
			refName string
		)

		// tags refs in tag events have "refs/tags/" prefix
		if tag, found := strings.CutPrefix(e.Ref, "refs/tags/"); found {
			refName = tag
		} else {
			log.WithContext(ctx).
				WithFields(log.Fields{
					"project-name": e.Project.Name,
					"ref":          e.Ref,
				}).
				Error("extracting tag name from ref")

			return
		}

		_ = deleteRef(ctx, c.Store, schemas.NewRef(
			schemas.NewProject(e.Project.PathWithNamespace),
			refKind,
			refName,
		), "received tag deletion tag event from webhook")
	}
}

func (c *Controller) processMergeEvent(ctx context.Context, e goGitlab.MergeEvent) {
	ref := schemas.NewRef(
		schemas.NewProject(e.Project.PathWithNamespace),
		schemas.RefKindMergeRequest,
		strconv.Itoa(e.ObjectAttributes.IID),
	)

	switch e.ObjectAttributes.Action {
	case "close":
		_ = deleteRef(ctx, c.Store, ref, "received merge request close event from webhook")
	case "merge":
		_ = deleteRef(ctx, c.Store, ref, "received merge request merge event from webhook")
	default:
		log.
			WithField("merge-request-event-type", e.ObjectAttributes.Action).
			Debug("received a non supported merge-request event type as a webhook")
	}
}

func (c *Controller) triggerRefMetricsPull(ctx context.Context, ref schemas.Ref) {
	logFields := log.Fields{
		"project-name": ref.Project.Name,
		"ref":          ref.Name,
		"ref-kind":     ref.Kind,
	}

	refExists, err := c.Store.RefExists(ctx, ref.Key())
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			WithError(err).
			Error("reading ref from the store")

		return
	}

	// Let's try to see if the project is configured to export this ref
	if !refExists {
		p := schemas.NewProject(ref.Project.Name)

		projectExists, err := c.Store.ProjectExists(ctx, p.Key())
		if err != nil {
			log.WithContext(ctx).
				WithFields(logFields).
				WithError(err).
				Error("reading project from the store")

			return
		}

		// Perhaps the project is discoverable through a wildcard
		if !projectExists && len(c.Config.Wildcards) > 0 {
			for _, w := range c.Config.Wildcards {
				// If in all our wildcards we have one which can potentially match the project ref
				// received, we trigger a pull of the project
				matches, err := isRefMatchingWilcard(w, ref)
				if err != nil {
					log.WithContext(ctx).
						WithError(err).
						Warn("checking if the ref matches the wildcard config")

					continue
				}

				if matches {
					c.ScheduleTask(context.TODO(), schemas.TaskTypePullProject, ref.Project.Name, ref.Project.Name, w.Pull)
					log.WithFields(logFields).Info("project ref not currently exported but its configuration matches a wildcard, triggering a pull of the project")
				} else {
					log.WithFields(logFields).Debug("project ref not matching wildcard, skipping..")
				}
			}

			log.WithFields(logFields).Info("done looking up for wildcards matching the project ref")

			return
		}

		if projectExists {
			// If the project exists, we check that the ref matches it's configuration
			if err := c.Store.GetProject(ctx, &p); err != nil {
				log.WithContext(ctx).
					WithFields(logFields).
					WithError(err).
					Error("reading project from the store")

				return
			}

			matches, err := isRefMatchingProjectPullRefs(p.Pull.Refs, ref)
			if err != nil {
				log.WithContext(ctx).
					WithError(err).
					Error("checking if the ref matches the project config")

				return
			}

			if matches {
				ref.Project = p

				if err = c.Store.SetRef(ctx, ref); err != nil {
					log.WithContext(ctx).
						WithFields(logFields).
						WithError(err).
						Error("writing ref in the store")

					return
				}

				goto schedulePull
			}
		}

		log.WithFields(logFields).Info("ref not configured in the exporter, ignoring pipeline webhook")

		return
	}

schedulePull:
	log.WithFields(logFields).Info("received a pipeline webhook from GitLab for a ref, triggering metrics pull")
	// TODO: When all the metrics will be sent over the webhook, we might be able to avoid redoing a pull
	// eg: 'coverage' is not in the pipeline payload yet, neither is 'artifacts' in the job one
	c.ScheduleTask(context.TODO(), schemas.TaskTypePullRefMetrics, string(ref.Key()), ref)
}

func (c *Controller) processDeploymentEvent(ctx context.Context, e goGitlab.DeploymentEvent) {
	c.triggerEnvironmentMetricsPull(
		ctx,
		schemas.Environment{
			ProjectName: e.Project.PathWithNamespace,
			Name:        e.Environment,
		},
	)
}

func (c *Controller) triggerEnvironmentMetricsPull(ctx context.Context, env schemas.Environment) {
	logFields := log.Fields{
		"project-name":     env.ProjectName,
		"environment-name": env.Name,
	}

	envExists, err := c.Store.EnvironmentExists(ctx, env.Key())
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			WithError(err).
			Error("reading environment from the store")

		return
	}

	if !envExists {
		p := schemas.NewProject(env.ProjectName)

		projectExists, err := c.Store.ProjectExists(ctx, p.Key())
		if err != nil {
			log.WithContext(ctx).
				WithFields(logFields).
				WithError(err).
				Error("reading project from the store")

			return
		}

		// Perhaps the project is discoverable through a wildcard
		if !projectExists && len(c.Config.Wildcards) > 0 {
			for _, w := range c.Config.Wildcards {
				// If in all our wildcards we have one which can potentially match the env
				// received, we trigger a pull of the project
				matches, err := isEnvMatchingWilcard(w, env)
				if err != nil {
					log.WithContext(ctx).
						WithError(err).
						Warn("checking if the env matches the wildcard config")

					continue
				}

				if matches {
					c.ScheduleTask(context.TODO(), schemas.TaskTypePullProject, env.ProjectName, env.ProjectName, w.Pull)
					log.WithFields(logFields).Info("project environment not currently exported but its configuration matches a wildcard, triggering a pull of the project")
				} else {
					log.WithFields(logFields).Debug("project ref not matching wildcard, skipping..")
				}
			}

			log.WithFields(logFields).Info("done looking up for wildcards matching the project ref")

			return
		}

		if projectExists {
			if err := c.Store.GetProject(ctx, &p); err != nil {
				log.WithContext(ctx).
					WithFields(logFields).
					WithError(err).
					Error("reading project from the store")
			}

			matches, err := isEnvMatchingProjectPullEnvironments(p.Pull.Environments, env)
			if err != nil {
				log.WithContext(ctx).
					WithError(err).
					Error("checking if the env matches the project config")

				return
			}

			if matches {
				// As we do not get the environment ID within the deployment event, we need to query it back..
				if err = c.UpdateEnvironment(ctx, &env); err != nil {
					log.WithContext(ctx).
						WithFields(logFields).
						WithError(err).
						Error("updating event from GitLab API")

					return
				}

				goto schedulePull
			}
		}

		log.WithFields(logFields).
			Info("environment not configured in the exporter, ignoring deployment webhook")

		return
	}

	// Need to refresh the env from the store in order to get at least it's ID
	if env.ID == 0 {
		if err = c.Store.GetEnvironment(ctx, &env); err != nil {
			log.WithContext(ctx).
				WithFields(logFields).
				WithError(err).
				Error("reading environment from the store")
		}
	}

schedulePull:
	log.WithFields(logFields).Info("received a deployment webhook from GitLab for an environment, triggering metrics pull")
	c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), env)
}

func isRefMatchingProjectPullRefs(pprs config.ProjectPullRefs, ref schemas.Ref) (matches bool, err error) {
	// We check if the ref kind is enabled
	switch ref.Kind {
	case schemas.RefKindBranch:
		if !pprs.Branches.Enabled {
			return
		}
	case schemas.RefKindTag:
		if !pprs.Tags.Enabled {
			return
		}
	case schemas.RefKindMergeRequest:
		if !pprs.MergeRequests.Enabled {
			return
		}
	default:
		return false, fmt.Errorf("invalid ref kind %v", ref.Kind)
	}

	// Then we check if it matches the regexp
	var re *regexp.Regexp

	if re, err = schemas.GetRefRegexp(pprs, ref.Kind); err != nil {
		return
	}

	return re.MatchString(ref.Name), nil
}

func isEnvMatchingProjectPullEnvironments(ppe config.ProjectPullEnvironments, env schemas.Environment) (matches bool, err error) {
	// We check if the environments pulling is enabled
	if !ppe.Enabled {
		return
	}

	// Then we check if it matches the regexp
	var re *regexp.Regexp

	if re, err = regexp.Compile(ppe.Regexp); err != nil {
		return
	}

	return re.MatchString(env.Name), nil
}

func isRefMatchingWilcard(w config.Wildcard, ref schemas.Ref) (matches bool, err error) {
	// Then we check if the owner matches the ref or is global
	if w.Owner.Kind != "" && !strings.Contains(ref.Project.Name, w.Owner.Name) {
		return
	}

	// Then we check if the ref matches the project pull parameters
	return isRefMatchingProjectPullRefs(w.Pull.Refs, ref)
}

func isEnvMatchingWilcard(w config.Wildcard, env schemas.Environment) (matches bool, err error) {
	// Then we check if the owner matches the ref or is global
	if w.Owner.Kind != "" && !strings.Contains(env.ProjectName, w.Owner.Name) {
		return
	}

	// Then we check if the ref matches the project pull parameters
	return isEnvMatchingProjectPullEnvironments(w.Pull.Environments, env)
}
