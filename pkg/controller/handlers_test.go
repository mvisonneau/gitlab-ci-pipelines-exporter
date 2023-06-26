package controller

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
)

func TestWebhookHandler(t *testing.T) {
	_, c, _, srv := newTestController(config.Config{
		Server: config.Server{
			Webhook: config.ServerWebhook{
				Enabled:     true,
				SecretToken: "secret",
			},
		},
	})
	srv.Close()

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	// Test without auth token, should return a 403
	w := httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	// Provide correct authentication header
	req.Header.Add("X-Gitlab-Token", "secret")

	// Test with empty body, should return a 400
	w = httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Provide an invalid body
	req.Body = ioutil.NopCloser(strings.NewReader(`[`))

	// Test with invalid body, should return a 400
	w = httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	// Provide an invalid event type
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "wiki_page"}`))
	req.Header.Set("X-Gitlab-Event", "Wiki Page Hook")

	// Test with invalid event type, should return a 422
	w = httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)

	// Provide an valid event type: pipeline
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "pipeline"}`))
	req.Header.Set("X-Gitlab-Event", "Pipeline Hook")

	// Test with pipeline event type, should return a 200
	w = httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// Provide an valid event type: deployment
	req.Body = ioutil.NopCloser(strings.NewReader(`{"object_kind": "deployment"}`))
	req.Header.Set("X-Gitlab-Event", "Deployment Hook")

	// Test with deployment event type, should return a 200
	w = httptest.NewRecorder()
	c.WebhookHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
