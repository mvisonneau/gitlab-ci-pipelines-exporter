package cmd

import (
	"fmt"
	"io/ioutil"
	"runtime"

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

	// Interval in seconds at which to poll projects from wildcards
	ProjectsPollingIntervalSeconds int `yaml:"projects_polling_interval_seconds"`

	// Interval in seconds to fetch refs from projects
	RefsPollingIntervalSeconds int `yaml:"refs_polling_interval_seconds"`

	// Interval in seconds to get new pipelines from refs (exponentially backing of to maximum value)
	PipelinesPollingIntervalSeconds int `yaml:"pipelines_polling_interval_seconds"`

	// Whether to attempt retrieving refs from pipelines when the exporter starts
	OnInitFetchRefsFromPipelines bool `yaml:"on_init_fetch_refs_from_pipelines"`

	// Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
	OnInitFetchRefsFromPipelinesDepthLimit int `yaml:"on_init_fetch_refs_from_pipelines_depth_limit"`

	// Sets the parallelism for polling projects from the API
	MaximumProjectsPollingWorkers int `yaml:"maximum_projects_poller_workers"`

	// Disable OpenMetrics content encoding in prometheus HTTP handler (default: false)
	DisableOpenmetricsEncoding bool `yaml:"disable_openmetrics_encoding"`

	// Default parameters which can be overridden at either the Project or Wildcard level
	Defaults Parameters `yaml:"defaults"`

	// List of projects to poll
	Projects []Project `yaml:"projects"`

	// List of wildcards to search projects from
	Wildcards []Wildcard `yaml:"wildcards"`
}

// Parameters for the fetching configuration of Projects and Wildcards
type Parameters struct {
	// Whether to attempt to retrieve job metrics from polled pipelines
	FetchPipelineJobMetricsValue *bool `yaml:"fetch_pipeline_job_metrics"`

	// Whether to report all pipeline / job statuses, or only report the one from the last job.
	OutputSparseStatusMetricsValue *bool `yaml:"output_sparse_status_metrics"`

	// Whether to attempt to retrieve variables included in the pipeline execution
	FetchPipelineVariablesValue *bool `yaml:"fetch_pipeline_variables"`

	// Regular expression to filter pipeline variables values to fetch (defaults to '.*')
	PipelineVariablesRegexpValue *string `yaml:"pipeline_variables_regexp"`

	// Regular expression to filter project refs to fetch (defaults to '.*')
	RefsRegexpValue *string `yaml:"refs_regexp"`
}

// Project holds information about a GitLab project
type Project struct {
	Parameters `yaml:",inline"`

	Name string `yaml:"name"`
}

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
	Parameters `yaml:",inline"`

	Search string `yaml:"search"`
	Owner  struct {
		Name             string `yaml:"name"`
		Kind             string `yaml:"kind"`
		IncludeSubgroups bool   `yaml:"include_subgroups"`
	} `yaml:"owner"`
	Archived bool `yaml:"archived"`
}

// Default values
const (
	defaultMaximumGitLabAPIRequestsPerSecond      = 10
	defaultOnInitFetchRefsFromPipelinesDepthLimit = 100
	defaultPipelinesPollingIntervalSeconds        = 30
	defaultProjectsPollingIntervalSeconds         = 1800
	defaultRefsPollingIntervalSeconds             = 300

	defaultFetchPipelineJobMetrics   = false
	defaultOutputSparseStatusMetrics = false
	defaultFetchPipelineVariables    = false
	defaultRefsRegexp                = `^master$`
	defaultPipelineVariablesRegexp   = `.*`

	errNoProjectsOrWildcardConfigured = "you need to configure at least one project/wildcard to poll, none given"
	errConfigFileNotFound             = "couldn't open config file : %v"
)

var cfg = &Config{}

// FetchPipelineJobMetrics ...
func (p *Project) FetchPipelineJobMetrics(cfg *Config) bool {
	if p.FetchPipelineJobMetricsValue != nil {
		return *p.FetchPipelineJobMetricsValue
	}

	if cfg.Defaults.FetchPipelineJobMetricsValue != nil {
		return *cfg.Defaults.FetchPipelineJobMetricsValue
	}

	return defaultFetchPipelineJobMetrics
}

// OutputSparseStatusMetrics ...
func (p *Project) OutputSparseStatusMetrics(cfg *Config) bool {
	if p.OutputSparseStatusMetricsValue != nil {
		return *p.OutputSparseStatusMetricsValue
	}

	if cfg.Defaults.OutputSparseStatusMetricsValue != nil {
		return *cfg.Defaults.OutputSparseStatusMetricsValue
	}

	return defaultOutputSparseStatusMetrics
}

// FetchPipelineVariables ...
func (p *Project) FetchPipelineVariables(cfg *Config) bool {
	if p.FetchPipelineVariablesValue != nil {
		return *p.FetchPipelineVariablesValue
	}

	if cfg.Defaults.FetchPipelineVariablesValue != nil {
		return *cfg.Defaults.FetchPipelineVariablesValue
	}

	return defaultFetchPipelineVariables
}

// PipelineVariablesRegexp ...
func (p *Project) PipelineVariablesRegexp(cfg *Config) string {
	if p.PipelineVariablesRegexpValue != nil {
		return *p.PipelineVariablesRegexpValue
	}

	if cfg.Defaults.PipelineVariablesRegexpValue != nil {
		return *cfg.Defaults.PipelineVariablesRegexpValue
	}

	return defaultPipelineVariablesRegexp
}

// RefsRegexp ...
func (p *Project) RefsRegexp(cfg *Config) string {
	if p.RefsRegexpValue != nil {
		return *p.RefsRegexpValue
	}

	if cfg.Defaults.RefsRegexpValue != nil {
		return *cfg.Defaults.RefsRegexpValue
	}

	return defaultRefsRegexp
}

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

	if cfg.ProjectsPollingIntervalSeconds == 0 {
		cfg.ProjectsPollingIntervalSeconds = defaultProjectsPollingIntervalSeconds
	}

	if cfg.RefsPollingIntervalSeconds == 0 {
		cfg.RefsPollingIntervalSeconds = defaultRefsPollingIntervalSeconds
	}

	if cfg.PipelinesPollingIntervalSeconds == 0 {
		cfg.PipelinesPollingIntervalSeconds = defaultPipelinesPollingIntervalSeconds
	}

	if cfg.OnInitFetchRefsFromPipelinesDepthLimit == 0 {
		cfg.OnInitFetchRefsFromPipelinesDepthLimit = defaultOnInitFetchRefsFromPipelinesDepthLimit
	}

	if cfg.Gitlab.URL == "" {
		cfg.Gitlab.URL = "https://gitlab.com"
	}

	if cfg.Gitlab.HealthURL == "" {
		cfg.Gitlab.HealthURL = fmt.Sprintf("%s/users/sign_in", cfg.Gitlab.URL)
	}

	if cfg.MaximumProjectsPollingWorkers == 0 {
		cfg.MaximumProjectsPollingWorkers = runtime.GOMAXPROCS(0)
	}

	return nil
}
