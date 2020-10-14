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
	"github.com/xanzy/go-gitlab"
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
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "{\"error\": \"invalid token\"")
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Warn("unable to read body of a received webhook")
		return
	}

	event, err := goGitlab.ParseHook(goGitlab.HookEventType(r), payload)
	if err != nil {
		log.WithFields(logFields).WithFields(logFields).WithField("error", err.Error()).Warn("unable to parse body of a received webhook")
		return
	}

	switch event := event.(type) {
	case *gitlab.PipelineEvent:
		processPipelineEvent(*event)
	default:
		log.WithFields(logFields).WithField("event-type", reflect.TypeOf(event).String()).Warn("received a non supported event type as a webhook")
	}
}

func triggerProjectRefMetricsPull(pr schemas.ProjectRef) {
	logFields := log.Fields{
		"project-id":   pr.ID,
		"project-name": pr.PathWithNamespace,
		"project-ref":  pr.Ref,
	}

	exists, err := store.ProjectRefExists(pr.Key())
	if err != nil {
		log.WithFields(logFields).WithField("error", err.Error()).Error("reading project ref from the store")
	}

	if !exists {
		projects, err := store.Projects()
		if err != nil {
			log.WithFields(logFields).WithField("error", err.Error()).Error("reading projects from the store")
		}

		for _, p := range projects {
			if p.Name == pr.PathWithNamespace {
				if regexp.MustCompile(p.Pull.Refs.Regexp()).MatchString(pr.Ref) {
					if err = store.SetProjectRef(pr); err != nil {
						log.WithFields(logFields).WithField("error", err.Error()).Error("writing project ref in the store")
					}
					goto schedulePull
				}
			}
		}

		log.WithFields(logFields).Info("project ref not configured in the exporter, ignoring pipeline hook")
		return
	}

schedulePull:
	log.WithFields(logFields).Info("received a pipeline webhook from GitLab for a project ref, triggering metrics pull")
	// TODO: When all the metrics will be sent over the webhook, we might be able to avoid redoing a pull
	// eg: 'coverage' is not in the pipeline payload yet, neither is 'artifacts' in the job one
	go schedulePullProjectRefMetrics(context.Background(), pr)
}

func processPipelineEvent(e goGitlab.PipelineEvent) {
	triggerProjectRefMetricsPull(schemas.ProjectRef{
		ID:                e.Project.ID,
		PathWithNamespace: e.Project.PathWithNamespace,
		Ref:               e.ObjectAttributes.Ref,
	})
}
