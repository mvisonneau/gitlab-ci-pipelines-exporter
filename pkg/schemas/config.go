package schemas

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Default values
const (
	defaultPullConfigMaximumGitLabAPIRequestsPerSecond                       = 10
	defaultPullConfigMetricsOnInit                                           = true
	defaultPullConfigMetricsScheduled                                        = true
	defaultPullConfigMetricsIntervalSeconds                                  = 30
	defaultPullConfigProjectsFromWildcardsOnInit                             = true
	defaultPullConfigProjectsFromWildcardsScheduled                          = true
	defaultPullConfigProjectsFromWildcardsIntervalSeconds                    = 1800
	defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsOnInit          = true
	defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsScheduled       = true
	defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsIntervalSeconds = 300

	errNoProjectsOrWildcardConfigured = "you need to configure at least one project/wildcard to poll, none given"
	errConfigFileNotFound             = "couldn't open config file : %v"
)

// Config represents what can be defined as a yaml config file
type Config struct {
	// GitLab configuration
	Gitlab GitlabConfig `yaml:"gitlab"`

	// Disable OpenMetrics content encoding in prometheus HTTP handler (default: false)
	DisableOpenmetricsEncoding bool `yaml:"disable_openmetrics_encoding"`

	// Pull configuration
	Pull PullConfig `yaml:"pull"`

	// Default parameters which can be overridden at either the Project or Wildcard level
	ProjectDefaults ProjectParameters `yaml:"project_defaults"`

	// List of projects to poll
	Projects []Project `yaml:"projects"`

	// List of wildcards to search projects from
	Wildcards Wildcards `yaml:"wildcards"`
}

// GitlabConfig ..
type GitlabConfig struct {
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
}

// PullConfig ..
type PullConfig struct {
	// Maximum amount of requests per seconds to make against the GitLab API (default: 10)
	MaximumGitLabAPIRequestsPerSecondValue *int `yaml:"maximum_gitlab_api_requests_per_second"`

	// PullMetrics configuration
	Metrics PullConfigMetrics `yaml:"metrics"`

	// ProjectsFromWildcards configuration
	ProjectsFromWildcards PullConfigProjectsFromWildcards `yaml:"projects_from_wildcards"`

	// ProjectRefsFromBranchesTagsMergeRequests configuration
	ProjectRefsFromBranchesTagsMergeRequests PullConfigProjectRefsFromBranchesTagsMergeRequests `yaml:"project_refs_from_branches_tags_and_mrs"`
}

// PullConfigParameters ..
type PullConfigParameters struct {
	OnInitValue          *bool `yaml:"on_init"`
	ScheduledValue       *bool `yaml:"scheduled"`
	IntervalSecondsValue *int  `yaml:"interval_seconds"`
}

// PullConfigMetrics ..
type PullConfigMetrics PullConfigParameters

// PullConfigProjectsFromWildcards ..
type PullConfigProjectsFromWildcards PullConfigParameters

// PullConfigProjectRefsFromBranchesTagsMergeRequests ..
type PullConfigProjectRefsFromBranchesTagsMergeRequests PullConfigParameters

// ParseConfig loads a yaml file into a Config structure
func ParseConfig(path string, cfg *Config) error {
	configFile, err := ioutil.ReadFile(filepath.Clean(path))
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

	// Configure GitLab parameters
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

// MaximumGitLabAPIRequestsPerSecond ...
func (pc *PullConfig) MaximumGitLabAPIRequestsPerSecond() int {
	if pc.MaximumGitLabAPIRequestsPerSecondValue != nil {
		return *pc.MaximumGitLabAPIRequestsPerSecondValue
	}

	return defaultPullConfigMaximumGitLabAPIRequestsPerSecond
}

// OnInit ..
func (pc *PullConfigMetrics) OnInit() bool {
	if pc.OnInitValue != nil {
		return *pc.OnInitValue
	}

	return defaultPullConfigMetricsOnInit
}

// Scheduled ..
func (pc *PullConfigMetrics) Scheduled() bool {
	if pc.ScheduledValue != nil {
		return *pc.ScheduledValue
	}

	return defaultPullConfigMetricsScheduled
}

// IntervalSeconds ..
func (pc *PullConfigMetrics) IntervalSeconds() int {
	if pc.IntervalSecondsValue != nil {
		return *pc.IntervalSecondsValue
	}

	return defaultPullConfigMetricsIntervalSeconds
}

// OnInit ..
func (pc *PullConfigProjectsFromWildcards) OnInit() bool {
	if pc.OnInitValue != nil {
		return *pc.OnInitValue
	}

	return defaultPullConfigProjectsFromWildcardsOnInit
}

// Scheduled ..
func (pc *PullConfigProjectsFromWildcards) Scheduled() bool {
	if pc.ScheduledValue != nil {
		return *pc.ScheduledValue
	}

	return defaultPullConfigProjectsFromWildcardsScheduled
}

// IntervalSeconds ..
func (pc *PullConfigProjectsFromWildcards) IntervalSeconds() int {
	if pc.IntervalSecondsValue != nil {
		return *pc.IntervalSecondsValue
	}

	return defaultPullConfigProjectsFromWildcardsIntervalSeconds
}

// OnInit ..
func (pc *PullConfigProjectRefsFromBranchesTagsMergeRequests) OnInit() bool {
	if pc.OnInitValue != nil {
		return *pc.OnInitValue
	}

	return defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsOnInit
}

// Scheduled ..
func (pc *PullConfigProjectRefsFromBranchesTagsMergeRequests) Scheduled() bool {
	if pc.ScheduledValue != nil {
		return *pc.ScheduledValue
	}

	return defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsScheduled
}

// IntervalSeconds ..
func (pc *PullConfigProjectRefsFromBranchesTagsMergeRequests) IntervalSeconds() int {
	if pc.IntervalSecondsValue != nil {
		return *pc.IntervalSecondsValue
	}

	return defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsIntervalSeconds
}
