package controller

import (
	"context"
	"regexp"
	"strings"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func (c *Controller) processPipelineEvent(e goGitlab.PipelineEvent) {
	var k schemas.RefKind
	if e.MergeRequest.IID != 0 {
		k = schemas.RefKindMergeRequest
	} else if e.ObjectAttributes.Tag {
		k = schemas.RefKindTag
	} else {
		k = schemas.RefKindBranch
	}

	c.triggerRefMetricsPull(schemas.Ref{
		Kind:        k,
		ProjectName: e.Project.PathWithNamespace,
		Name:        e.ObjectAttributes.Ref,
	})
}

func (c *Controller) triggerRefMetricsPull(ref schemas.Ref) {
	logFields := log.Fields{
		"project-name": ref.ProjectName,
		"ref":          ref.Name,
		"ref-kind":     ref.Kind,
	}

	exists, err := c.Store.RefExists(ref.Key())
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Error("reading ref from the store")
	}

	// Let's try to see if the project is configured to export this ref
	if !exists {
		p := config.Project{
			Name: ref.ProjectName,
		}

		exists, err = c.Store.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
		}

		// Perhaps the project is discoverable through a wildcard
		if !exists && len(c.Wildcards) > 0 {
			for _, w := range c.Wildcards {
				// If in all our wildcards we have one which can potentially match the project ref
				// received, we trigger a scan
				if w.Owner.Kind == "" ||
					(strings.Contains(p.Name, w.Owner.Name) && regexp.MustCompile(w.Pull.Refs.Regexp).MatchString(ref.Name)) {
					c.ScheduleTask(context.TODO(), TaskTypePullProjectsFromWildcard, w)
					log.WithFields(logFields).Info("project ref not currently exported but its configuration matches a wildcard, triggering a pull of the projects from this wildcard")
					return
				}
			}
		}

		if exists {
			if err := c.Store.GetProject(&p); err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
			}

			if regexp.MustCompile(p.Pull.Refs.Regexp).MatchString(ref.Name) {
				if err = c.Store.SetRef(ref); err != nil {
					log.WithFields(logFields).WithField("error", err.Error()).Error("writing ref in the store")
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
	c.ScheduleTask(context.TODO(), TaskTypePullRefMetrics, ref)
}

func (c *Controller) processDeploymentEvent(e goGitlab.DeploymentEvent) {
	c.triggerEnvironmentMetricsPull(schemas.Environment{
		ProjectName: e.Project.PathWithNamespace,
		Name:        e.Environment,
	})
}

func (c *Controller) triggerEnvironmentMetricsPull(env schemas.Environment) {
	logFields := log.Fields{
		"project-name":     env.ProjectName,
		"environment-name": env.Name,
	}

	exists, err := c.Store.EnvironmentExists(env.Key())
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Error("reading environment from the store")
	}

	if !exists {
		p := config.Project{
			Name: env.ProjectName,
		}

		exists, err = c.Store.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
		}

		// Perhaps the project is discoverable through a wildcard
		if !exists && len(c.Wildcards) > 0 {
			for _, w := range c.Wildcards {
				// If in all our wildcards we have one which can potentially match the project ref
				// received, we trigger a scan
				if w.Pull.Environments.Enabled && (w.Owner.Kind == "" || (strings.Contains(p.Name, w.Owner.Name) && regexp.MustCompile(w.Pull.Environments.Regexp).MatchString(env.ProjectName))) {
					c.ScheduleTask(context.TODO(), TaskTypePullProjectsFromWildcard, w)
					log.WithFields(logFields).Info("project environment not currently exported but its configuration matches a wildcard, triggering a pull of the projects from this wildcard")
					return
				}
			}
		}

		if exists {
			if err := c.Store.GetProject(&p); err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
			}

			// As we do not get the environment ID within the deployment event, we need to query it back..
			envs, err := c.Gitlab.GetProjectEnvironments(p.Name, p.Pull.Environments.Regexp)
			if err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Error("listing project envs from GitLab API")
			}

			for envID, envName := range envs {
				if envName == env.Name {
					env.ID = envID
					break
				}
			}

			if env.ID != 0 {
				if err = c.Store.SetEnvironment(env); err != nil {
					log.WithFields(logFields).WithField("error", err.Error()).Error("writing environment in the store")
				}
				goto schedulePull
			}
		}

		log.WithFields(logFields).Info("environment not configured in the exporter, ignoring deployment webhook")
		return
	}

	// Need to refresh the env from the store in order to get at least it's ID
	if env.ID == 0 {
		if err = c.Store.GetEnvironment(&env); err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading environment from the store")
		}
	}

schedulePull:
	log.WithFields(logFields).Info("received a deployment webhook from GitLab for an environment, triggering metrics pull")
	c.ScheduleTask(context.TODO(), TaskTypePullEnvironmentMetrics, env)
}
