package exporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// WebhookHandler ..
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	logFields := log.Fields{
		"ip-address": r.RemoteAddr,
		"user-agent": r.UserAgent(),
	}
	log.WithFields(logFields).Debug("webhook request")

	if r.Header.Get("X-Gitlab-Token") != config.Server.Webhook.SecretToken {
		log.WithFields(logFields).Debug("invalid token provided for a webhook request")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"error\": \"invalid token\"")
		return
	}

	if r.Body == http.NoBody {
		log.WithFields(logFields).WithField("error", "nil body").Warn("unable to read body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Warn("unable to read body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := goGitlab.ParseHook(goGitlab.HookEventType(r), payload)
	if err != nil {
		log.WithFields(logFields).WithFields(logFields).WithField("error", err.Error()).Warn("unable to parse body of a received webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case *goGitlab.PipelineEvent:
		go processPipelineEvent(*event)
	case *goGitlab.DeploymentEvent:
		go processDeploymentEvent(*event)
	default:
		log.WithFields(logFields).WithField("event-type", reflect.TypeOf(event).String()).Warn("received a non supported event type as a webhook")
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func processPipelineEvent(e goGitlab.PipelineEvent) {
	var k schemas.RefKind
	if e.MergeRequest.IID != 0 {
		k = schemas.RefKindMergeRequest
	} else if e.ObjectAttributes.Tag {
		k = schemas.RefKindTag
	} else {
		k = schemas.RefKindBranch
	}

	triggerRefMetricsPull(schemas.Ref{
		Kind:        k,
		ProjectName: e.Project.PathWithNamespace,
		Name:        e.ObjectAttributes.Ref,
	})
}

func triggerRefMetricsPull(ref schemas.Ref) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	logFields := log.Fields{
		"project-name": ref.ProjectName,
		"ref":          ref.Name,
		"ref-kind":     ref.Kind,
	}

	exists, err := store.RefExists(ref.Key())
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Error("reading ref from the store")
	}

	if !exists {
		p := schemas.Project{
			Name: ref.ProjectName,
		}

		exists, err = store.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
		}

		if exists {
			if err := store.GetProject(&p); err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
			}

			if regexp.MustCompile(p.Pull.Refs.Regexp()).MatchString(ref.Name) {
				if err = store.SetRef(ref); err != nil {
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
	go schedulePullRefMetrics(context.Background(), ref)
}

func processDeploymentEvent(e goGitlab.DeploymentEvent) {
	triggerEnvironmentMetricsPull(schemas.Environment{
		ProjectName: e.Project.PathWithNamespace,
		Name:        e.Environment,
	})
}

func triggerEnvironmentMetricsPull(env schemas.Environment) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	logFields := log.Fields{
		"project-name":     env.ProjectName,
		"environment-name": env.Name,
	}

	exists, err := store.EnvironmentExists(env.Key())
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Error("reading environment from the store")
	}

	if !exists {
		p := schemas.Project{
			Name: env.ProjectName,
		}

		exists, err = store.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
		}

		if exists {
			if err := store.GetProject(&p); err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Error("reading project from the store")
			}

			// As we do not get the environment ID within the deployment event, we need to query it back..
			envs, err := gitlabClient.GetProjectEnvironments(p.Name, p.Pull.Environments.NameRegexp())
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
				if err = store.SetEnvironment(env); err != nil {
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
		if err = store.GetEnvironment(&env); err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading environment from the store")
		}
	}

schedulePull:
	log.WithFields(logFields).Info("received a deployment webhook from GitLab for an environment, triggering metrics pull")
	go schedulePullEnvironmentMetrics(context.Background(), env)
}
