package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"
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
	PipelinesMaxPollingIntervalSeconds     int        `yaml:"pipelines_max_polling_interval_seconds"`        // Maximum interval in seconds to fetch new pipelines from refs
	OnInitFetchRefsFromPipelines           bool       `yaml:"on_init_fetch_refs_from_pipelines"`             // Whether to attempt retrieving refs from pipelines when the exporter starts
	OnInitFetchRefsFromPipelinesDepthLimit int        `yaml:"on_init_fetch_refs_from_pipelines_depth_limit"` // Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
	DefaultRefsRegexp                      string     `yaml:"default_refs"`                                  // Default regular expression
	Projects                               []Project  `yaml:"projects"`                                      // List of projects to poll
	Wildcards                              []Wildcard `yaml:"wildcards"`                                     // List of wildcards to search projects from
}

// Project holds information about a GitLab project
type Project struct {
	Name string
	Refs string
}

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
	Search string
	Owner  struct {
		Name             string
		Kind             string
		IncludeSubgroups bool `yaml:"include_subgroups"`
	}
	Archived bool `yaml:"archived"`
	Refs     string
}

// Default values
const (
	defaultMaximumGitLabAPIRequestsPerSecond      = 10
	defaultOnInitFetchRefsFromPipelinesDepthLimit = 100
	defaultPipelinesMaxPollingIntervalSeconds     = 3600
	defaultPipelinesPollingIntervalSeconds        = 30
	defaultProjectsPollingIntervalSeconds         = 1800
	defaultRefsPollingIntervalSeconds             = 300
)

var cfg *Config

// Parse loads a yaml file into a Config structure
func (cfg *Config) Parse(path string) error {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Couldn't open config file : %s", err.Error())
	}

	err = yaml.Unmarshal(configFile, cfg)
	if err != nil {
		return fmt.Errorf("Unable to parse config file: %s", err.Error())
	}

	if len(cfg.Projects) < 1 && len(cfg.Wildcards) < 1 {
		return fmt.Errorf("You need to configure at least one project/wildcard to poll, none given")
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

	if cfg.PipelinesMaxPollingIntervalSeconds == 0 {
		cfg.PipelinesMaxPollingIntervalSeconds = defaultPipelinesMaxPollingIntervalSeconds
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
