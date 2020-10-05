package gitlab

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

// Mocking helpers
func getMockedGitlabClient() (*http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	opts := []goGitlab.ClientOptionFunc{
		goGitlab.WithBaseURL(server.URL),
		goGitlab.WithoutRetries(),
	}

	gc, _ := goGitlab.NewClient("", opts...)

	c := &Client{
		Client:      gc,
		RateLimiter: ratelimit.New(100),
	}

	return mux, server, c
}

func TestNewHTTPClient(t *testing.T) {
	expected := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	c := NewHTTPClient(true)
	assert.EqualValues(t, expected, c)
}
