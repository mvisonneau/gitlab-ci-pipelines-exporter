package exporter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	resetGlobalValues()

	config.Server.Webhook.SecretToken = "secret"
	req := httptest.NewRequest("POST", "/webhook", nil)

	// Test without auth token, should return a 403
	w := httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	// Provide correct authentication header
	req.Header.Add("X-Gitlab-Token", "secret")

	// Test with empty body, should return a 400
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Provide an invalid body
	req.Body = ioutil.NopCloser(strings.NewReader(`[`))

	// Test with invalid body, should return a 400
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Provide an invalid event type
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "wiki_page"}`))
	req.Header.Set("X-Gitlab-Event", "Wiki Page Hook")

	// Test with invalid event type, should return a 422
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)

	// Provide an valid event type: pipeline
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "pipeline"}`))
	req.Header.Set("X-Gitlab-Event", "Pipeline Hook")

	// Test with pipeline event type, should return a 200
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// Provide an valid event type: deployment
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "deployment"}`))
	req.Header.Set("X-Gitlab-Event", "Deployment Hook")

	// Test with deployment event type, should return a 200
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestTriggerRefMetricsPull(_ *testing.T) {
	resetGlobalValues()

	ref1 := schemas.Ref{
		ProjectName: "group/foo",
		Name:        "main",
	}

	p2 := schemas.Project{Name: "group/bar"}
	ref2 := schemas.Ref{
		ProjectName: "group/bar",
		Name:        "main",
	}

	store.SetRef(ref1)
	store.SetProject(p2)

	// TODO: Assert results somehow
	triggerRefMetricsPull(ref1)
	triggerRefMetricsPull(ref2)
}

func TestTriggerEnvironmentMetricsPull(_ *testing.T) {
	resetGlobalValues()

	p1 := schemas.Project{Name: "foo/bar"}
	env1 := schemas.Environment{
		ProjectName: "foo/bar",
		Name:        "dev",
	}

	env2 := schemas.Environment{
		ProjectName: "foo/baz",
		Name:        "prod",
	}

	store.SetProject(p1)
	store.SetEnvironment(env1)
	store.SetEnvironment(env2)

	// TODO: Assert results somehow
	triggerEnvironmentMetricsPull(env1)
	triggerEnvironmentMetricsPull(env2)
}
