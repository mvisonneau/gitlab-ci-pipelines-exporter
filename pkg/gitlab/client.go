package gitlab

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/paulbellamy/ratecounter"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
)

const (
	userAgent  = "gitlab-ci-pipelines-exporter"
	tracerName = "gitlab-ci-pipelines-exporter"
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
	RequestsCounter   atomic.Uint64
	RequestsLimit     int
	RequestsRemaining int

	version GitLabVersion
	mutex   sync.RWMutex
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
func (c *Client) ReadinessCheck(ctx context.Context) healthcheck.Check {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ReadinessCheck")
	defer span.End()

	return func() error {
		if c.Readiness.HTTPClient == nil {
			return fmt.Errorf("readiness http client not configured")
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			c.Readiness.URL,
			nil,
		)
		if err != nil {
			return err
		}

		resp, err := c.Readiness.HTTPClient.Do(req)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("HTTP error: empty response")
		}

		if err == nil && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		return nil
	}
}

func (c *Client) rateLimit(ctx context.Context) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:rateLimit")
	defer span.End()

	ratelimit.Take(ctx, c.RateLimiter)
	// Used for monitoring purposes
	c.RateCounter.Incr(1)
	c.RequestsCounter.Add(1)
}

func (c *Client) UpdateVersion(version GitLabVersion) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.version = version
}

func (c *Client) Version() GitLabVersion {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.version
}

func (c *Client) requestsRemaining(response *goGitlab.Response) {
	if response == nil {
		return
	}

	if remaining := response.Header.Get("ratelimit-remaining"); remaining != "" {
		c.RequestsRemaining, _ = strconv.Atoi(remaining)
	}

	if limit := response.Header.Get("ratelimit-limit"); limit != "" {
		c.RequestsLimit, _ = strconv.Atoi(limit)
	}
}
