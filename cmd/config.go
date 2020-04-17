package cmd

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"
)

// Config represents what can be defined as a yaml config file
type Config struct {
	Gitlab struct {
		URL           string `yaml:"url"`             // The URL of the GitLab server/api (default to https://gitlab.com)
		Token         string `yaml:"token"`           // Token to use to authenticate against the API
		HealthURL     string `yaml:"health_url"`      // The URL of the GitLab server/api health endpoint (default to /users/sign_in which is publicly available on gitlab.com)
		SkipTLSVerify bool   `yaml:"skip_tls_verify"` // Whether to validate TLS certificates or not
	}

	MaximumGitLabAPIRequestsPerSecond      int        `yaml:"maximum_gitlab_api_requests_per_second"`        // Maximum amount of requests per seconds to make against the GitLab API (default: 10)
	ProjectsPollingIntervalSeconds         int        `yaml:"projects_polling_interval_seconds"`             // Interval in seconds at which to poll projects from wildcards
	RefsPollingIntervalSeconds             int        `yaml:"refs_polling_interval_seconds"`                 // Interval in seconds to fetch refs from projects
	PipelinesPollingIntervalSeconds        int        `yaml:"pipelines_polling_interval_seconds"`            // Interval in seconds to get new pipelines from refs (exponentially backing of to maximum value)
	FetchPipelineJobMetrics                bool       `yaml:"fetch_pipeline_job_metrics"`                    // Whether to attempt to retrieve job metrics from polled pipelines
	OutputSparseStatusMetrics              bool       `yaml:"output_sparse_status_metrics"`                  // Whether to report all pipeline / job statuses, or only report the one from the last job.
	OnInitFetchRefsFromPipelines           bool       `yaml:"on_init_fetch_refs_from_pipelines"`             // Whether to attempt retrieving refs from pipelines when the exporter starts
	OnInitFetchRefsFromPipelinesDepthLimit int        `yaml:"on_init_fetch_refs_from_pipelines_depth_limit"` // Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
	DefaultRefsRegexp                      string     `yaml:"default_refs"`                                  // Default regular expression
	MaximumProjectsPollingWorkers          int        `yaml:"maximum_projects_poller_workers"`               // Sets the parallelism for polling projects from the API
	PrometheusOpenmetricsEncoding          bool       `yaml:"prometheus_openmetrics_encoding"`               // Enable OpenMetrics content encoding in prometheus HTTP handler (default: false)
	Projects                               []Project  `yaml:"projects"`                                      // List of projects to poll
	Wildcards                              []Wildcard `yaml:"wildcards"`                                     // List of wildcards to search projects from
}

// Project holds information about a GitLab project
type Project struct {
	Name                      string `yaml:"name"`
	Refs                      string `yaml:"refs"`
	FetchPipelineJobMetrics   *bool  `yaml:"fetch_pipeline_job_metrics,omitempty"`
	OutputSparseStatusMetrics *bool  `yaml:"output_sparse_status_metrics,omitempty"`
	GitlabProject             *gitlab.Project
}

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
	Search string
	Owner  struct {
		Name             string
		Kind             string
		IncludeSubgroups bool `yaml:"include_subgroups"`
	}
	Archived                  bool  `yaml:"archived"`
	FetchPipelineJobMetrics   *bool `yaml:"fetch_pipeline_job_metrics,omitempty"`
	OutputSparseStatusMetrics *bool `yaml:"output_sparse_status_metrics,omitempty"`
	Refs                      string
}

// Default values
const (
	defaultMaximumGitLabAPIRequestsPerSecond      = 10
	defaultOnInitFetchRefsFromPipelinesDepthLimit = 100
	defaultPipelinesPollingIntervalSeconds        = 30
	defaultProjectsPollingIntervalSeconds         = 1800
	defaultRefsPollingIntervalSeconds             = 300

	errNoProjectsOrWildcardConfigured = "you need to configure at least one project/wildcard to poll, none given"
	errConfigFileNotFound             = "couldn't open config file : %v"
)

var cfg = &Config{}

// ShouldFetchPipelineJobMetrics returns true if pipeline job statistics should be fetched
func (p *Project) ShouldFetchPipelineJobMetrics(cfg *Config) bool {
	if p.FetchPipelineJobMetrics == nil {
		// Default to global config value
		return cfg.FetchPipelineJobMetrics
	}
	return *p.FetchPipelineJobMetrics
}

// ShouldOutputSparseStatusMetrics returns true if sparse status metrics should be exported
func (p *Project) ShouldOutputSparseStatusMetrics(cfg *Config) bool {
	if p.OutputSparseStatusMetrics == nil {
		// Default to global config value
		return cfg.OutputSparseStatusMetrics
	}
	return *p.OutputSparseStatusMetrics
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

// MergeWithContext is used to override values defined in the config by ones
// provided at runtime
func (cfg *Config) MergeWithContext(ctx *cli.Context) {
	token := ctx.GlobalString("gitlab-token")
	if len(token) != 0 {
		cfg.Gitlab.Token = token
	}
}
