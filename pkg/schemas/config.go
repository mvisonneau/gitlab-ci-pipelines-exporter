package schemas

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config represents what can be defined as a yaml config file
type Config struct {
	// GitLab configuration
	Gitlab struct {
		// The URL of the GitLab server/api (default to https://gitlab.com)
		URL string `yaml:"url"`

		// Token to use to authenticate against the API
		Token string `yaml:"token"`

		// The URL of the GitLab server/api health endpoint (default to /users/sign_in which is publicly available on gitlab.com)
		HealthURL string `yaml:"health_url"`

		// Whether to validate the service is reachable calling HealthURL
		DisableHealthCheck bool `yaml:"disable_health_check"`

		// Whether to skip TLS validation when querying HealthURL
		DisableTLSVerify bool `yaml:"disable_tls_verify"`
	} `yaml:"gitlab"`

	// Maximum amount of requests per seconds to make against the GitLab API (default: 10)
	MaximumGitLabAPIRequestsPerSecond int `yaml:"maximum_gitlab_api_requests_per_second"`

	// Interval in seconds to discover projects from wildcards
	WildcardsProjectsDiscoverIntervalSeconds int `yaml:"wildcards_projects_discover_interval_seconds"`

	// Interval in seconds to discover refs from projects
	ProjectsRefsDiscoverIntervalSeconds int `yaml:"projects_refs_discover_interval_seconds"`

	// Interval in seconds to poll metrics from discovered project refs
	ProjectsRefsPollingIntervalSeconds int `yaml:"projects_refs_polling_interval_seconds"`

	// Whether to attempt retrieving refs from pipelines when the exporter starts
	OnInitFetchRefsFromPipelines bool `yaml:"on_init_fetch_refs_from_pipelines"`

	// Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
	OnInitFetchRefsFromPipelinesDepthLimit int `yaml:"on_init_fetch_refs_from_pipelines_depth_limit"`

	// Disable OpenMetrics content encoding in prometheus HTTP handler (default: false)
	DisableOpenmetricsEncoding bool `yaml:"disable_openmetrics_encoding"`

	// Default parameters which can be overridden at either the Project or Wildcard level
	Defaults Parameters `yaml:"defaults"`

	// List of projects to poll
	Projects []Project `yaml:"projects"`

	// List of wildcards to search projects from
	Wildcards Wildcards `yaml:"wildcards"`
}

// Default values
const (
	defaultMaximumGitLabAPIRequestsPerSecond      = 10
	defaultOnInitFetchRefsFromPipelinesDepthLimit = 100

	defaultWildcardsProjectsDiscoverIntervalSeconds = 1800
	defaultProjectsRefsDiscoverIntervalSeconds      = 300
	defaultProjectsRefsPollingIntervalSeconds       = 30

	errNoProjectsOrWildcardConfigured = "you need to configure at least one project/wildcard to poll, none given"
	errConfigFileNotFound             = "couldn't open config file : %v"
)

// Parse loads a yaml file into a Config structure
func (cfg *Config) Parse(path string) error {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errConfigFileNotFound, err)
	}

	err = yaml.Unmarshal(configFile, cfg)
	if err != nil {
		return fmt.Errorf("unable to parse config file: %v", err)
	}

	if len(cfg.Projects) < 1 && len(cfg.Wildcards) < 1 {
		return fmt.Errorf(errNoProjectsOrWildcardConfigured)
	}

	// Defining defaults
	if cfg.MaximumGitLabAPIRequestsPerSecond == 0 {
		cfg.MaximumGitLabAPIRequestsPerSecond = defaultMaximumGitLabAPIRequestsPerSecond
	}

	if cfg.WildcardsProjectsDiscoverIntervalSeconds == 0 {
		cfg.WildcardsProjectsDiscoverIntervalSeconds = defaultWildcardsProjectsDiscoverIntervalSeconds
	}

	if cfg.ProjectsRefsDiscoverIntervalSeconds == 0 {
		cfg.ProjectsRefsDiscoverIntervalSeconds = defaultProjectsRefsDiscoverIntervalSeconds
	}

	if cfg.ProjectsRefsPollingIntervalSeconds == 0 {
		cfg.ProjectsRefsPollingIntervalSeconds = defaultProjectsRefsPollingIntervalSeconds
	}

	if cfg.OnInitFetchRefsFromPipelinesDepthLimit == 0 {
		cfg.OnInitFetchRefsFromPipelinesDepthLimit = defaultOnInitFetchRefsFromPipelinesDepthLimit
	}

	if cfg.Gitlab.URL == "" {
		cfg.Gitlab.URL = "https://gitlab.com"
	}

	if cfg.Gitlab.HealthURL == "" {
		// Hack to fix the missing health endpoint on gitlab.com
		if cfg.Gitlab.URL == "https://gitlab.com" {
			cfg.Gitlab.HealthURL = "https://gitlab.com/explore"
		} else {
			cfg.Gitlab.HealthURL = fmt.Sprintf("%s/-/health", cfg.Gitlab.URL)
		}
	}

	return nil
}
