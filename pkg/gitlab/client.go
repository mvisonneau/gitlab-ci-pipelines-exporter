package gitlab

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/xanzy/go-gitlab"
	goGitlab "github.com/xanzy/go-gitlab"
)

const (
	userAgent = "gitlab-ci-pipelines-exporter"
)

// Client ..
type Client struct {
	*goGitlab.Client

	Readiness struct {
		URL        string
		HTTPClient *http.Client
	}

	RateLimiter ratelimit.Limiter
}

// ClientConfig ..
type ClientConfig struct {
	URL              string
	Token            string
	UserAgentVersion string
	DisableTLSVerify bool
	ReadinessURL     string

	RateLimiter ratelimit.Limiter
}

// NewHTTPClient ..
func NewHTTPClient(disableTLSVerify bool) *http.Client {
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: disableTLSVerify},
	}}
}

// NewClient ..
func NewClient(cfg ClientConfig) (*Client, error) {
	opts := []goGitlab.ClientOptionFunc{
		gitlab.WithHTTPClient(NewHTTPClient(cfg.DisableTLSVerify)),
		gitlab.WithBaseURL(cfg.URL),
		gitlab.WithoutRetries(),
	}

	gc, err := gitlab.NewClient(cfg.Token, opts...)
	if err != nil {
		return nil, err
	}

	gc.UserAgent = fmt.Sprintf("%s-%s", userAgent, cfg.UserAgentVersion)

	readinessCheckHTTPClient := NewHTTPClient(cfg.DisableTLSVerify)
	readinessCheckHTTPClient.Timeout = 5 * time.Second

	return &Client{
		Client:      gc,
		RateLimiter: cfg.RateLimiter,
		Readiness: struct {
			URL        string
			HTTPClient *http.Client
		}{
			URL:        cfg.ReadinessURL,
			HTTPClient: readinessCheckHTTPClient,
		},
	}, nil
}

// ReadinessCheck ..
func (c *Client) ReadinessCheck() healthcheck.Check {
	return func() error {
		resp, err := c.Readiness.HTTPClient.Get(c.Readiness.URL)
		if err == nil && resp.StatusCode != 200 {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}
		return err
	}
}

func (c *Client) rateLimit() {
	ratelimit.Take(c.RateLimiter)
}
