// +build !race

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

	// Provide an valid event type
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "pipeline"}`))
	req.Header.Set("X-Gitlab-Event", "Pipeline Hook")

	// Test with invalid event type, should return a 200
	w = httptest.NewRecorder()
	WebhookHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestTriggerProjectRefMetricsPull(_ *testing.T) {
	resetGlobalValues()
	configureStore()

	pr1 := schemas.ProjectRef{
		ID:                1,
		PathWithNamespace: "group/foo",
		Ref:               "main",
	}

	p2 := schemas.Project{Name: "group/bar"}
	pr2 := schemas.ProjectRef{
		Project:           p2,
		ID:                2,
		PathWithNamespace: "group/bar",
		Ref:               "main",
	}

	store.SetProjectRef(pr1)
	store.SetProject(p2)

	// TODO: Assert results somehow
	triggerProjectRefMetricsPull(pr1)
	triggerProjectRefMetricsPull(pr2)
}
