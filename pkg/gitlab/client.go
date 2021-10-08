package gitlab

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/paulbellamy/ratecounter"
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

	RateLimiter       ratelimit.Limiter
	RateCounter       *ratecounter.RateCounter
	RequestsCounter   uint64
	RequestsLimit     int
	RequestsRemaining int
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
	// http.DefaultTransport contains useful settings such as the correct values for the picking
	// up proxy informations from env variables
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: disableTLSVerify}

	return &http.Client{
		Transport: transport,
	}
}

// NewClient ..
func NewClient(cfg ClientConfig) (*Client, error) {
	opts := []goGitlab.ClientOptionFunc{
		goGitlab.WithHTTPClient(NewHTTPClient(cfg.DisableTLSVerify)),
		goGitlab.WithBaseURL(cfg.URL),
		goGitlab.WithoutRetries(),
	}

	gc, err := goGitlab.NewOAuthClient(cfg.Token, opts...)
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
		RateCounter: ratecounter.NewRateCounter(time.Second),
	}, nil
}

// ReadinessCheck ..
func (c *Client) ReadinessCheck() healthcheck.Check {
	return func() error {
		if c.Readiness.HTTPClient == nil {
			return fmt.Errorf("readiness http client not configured")
		}

		resp, err := c.Readiness.HTTPClient.Get(c.Readiness.URL)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("HTTP error: empty response")
		}

		if err == nil && resp.StatusCode != 200 {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		return nil
	}
}

func (c *Client) rateLimit() {
	ratelimit.Take(c.RateLimiter)
	// Used for monitoring purposes
	c.RateCounter.Incr(1)
	c.RequestsCounter++
}

func (c *Client) requestsRemaining(response *goGitlab.Response) {
	remaining := response.Header.Get("ratelimit-remaining")
	if remaining != "" {
		c.RequestsRemaining, _ = strconv.Atoi(remaining)
	}
	limit := response.Header.Get("ratelimit-limit")
	if limit != "" {
		c.RequestsLimit, _ = strconv.Atoi(limit)
	}
}
