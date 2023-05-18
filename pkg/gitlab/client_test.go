package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paulbellamy/ratecounter"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
)

// Mocking helpers.
func getMockedClient() (context.Context, *http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	opts := []goGitlab.ClientOptionFunc{
		goGitlab.WithBaseURL(server.URL),
		goGitlab.WithoutRetries(),
	}

	gc, _ := goGitlab.NewClient("", opts...)

	c := &Client{
		Client:      gc,
		RateLimiter: ratelimit.NewLocalLimiter(100, 1),
		RateCounter: ratecounter.NewRateCounter(time.Second),
	}

	return context.Background(), mux, server, c
}

func TestNewHTTPClient(t *testing.T) {
	c := NewHTTPClient(true)
	assert.True(t, c.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
}

func TestNewClient(t *testing.T) {
	cfg := ClientConfig{
		URL:              "https://gitlab.example.com",
		Token:            "supersecret",
		UserAgentVersion: "0.0.0",
		DisableTLSVerify: true,
		ReadinessURL:     "https://gitlab.example.com/amialive",
		RateLimiter:      ratelimit.NewLocalLimiter(10, 1),
	}

	c, err := NewClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, c.Client)
	assert.Equal(t, "gitlab-ci-pipelines-exporter-0.0.0", c.Client.UserAgent)
	assert.Equal(t, "https", c.Client.BaseURL().Scheme)
	assert.Equal(t, "gitlab.example.com", c.Client.BaseURL().Host)
	assert.Equal(t, "https://gitlab.example.com/amialive", c.Readiness.URL)
	assert.True(t, c.Readiness.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
	assert.Equal(t, 5*time.Second, c.Readiness.HTTPClient.Timeout)
}

func TestReadinessCheck(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	mux.HandleFunc(
		"/200",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			w.WriteHeader(http.StatusOK)
		},
	)
	mux.HandleFunc(
		"/500",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)

	readinessCheck := c.ReadinessCheck(ctx)
	assert.Error(t, readinessCheck())

	c.Readiness.HTTPClient = NewHTTPClient(false)
	c.Readiness.URL = fmt.Sprintf("%s/200", server.URL)

	assert.NoError(t, readinessCheck())

	c.Readiness.URL = fmt.Sprintf("%s/500", server.URL)

	assert.Error(t, readinessCheck())
}
