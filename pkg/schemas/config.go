package schemas

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Default values
const (
	defaultServerConfigEnablePprof                                 = false
	defaultServerConfigListenAddress                               = ":8080"
	defaultServerConfigMetricsEnabled                              = true
	defaultServerConfigMetricsEnableOpenmetricsEncoding            = false
	defaultServerConfigWebhookEnabled                              = false
	defaultGitlabConfigURL                                         = "https://gitlab.com"
	defaultGitlabConfigHealthURL                                   = "https://gitlab.com/explore"
	defaultGitlabConfigEnableHealthCheck                           = true
	defaultGitlabConfigEnableTLSVerify                             = true
	defaultPullConfigMaximumGitLabAPIRequestsPerSecond             = 10
	defaultPullConfigProjectsFromWildcardsOnInit                   = true
	defaultPullConfigProjectsFromWildcardsScheduled                = true
	defaultPullConfigProjectsFromWildcardsIntervalSeconds          = 1800
	defaultPullConfigEnvironmentsFromProjectsOnInit                = true
	defaultPullConfigEnvironmentsFromProjectsScheduled             = true
	defaultPullConfigEnvironmentsFromProjectsIntervalSeconds       = 1800
	defaultPullConfigRefsFromProjectsOnInit                        = true
	defaultPullConfigRefsFromProjectsScheduled                     = true
	defaultPullConfigRefsFromProjectsIntervalSeconds               = 300
	defaultPullConfigMetricsOnInit                                 = true
	defaultPullConfigMetricsScheduled                              = true
	defaultPullConfigMetricsIntervalSeconds                        = 30
	defaultGarbageCollectConfigProjectsOnInit                      = false
	defaultGarbageCollectConfigProjectsScheduled                   = true
	defaultGarbageCollectConfigProjectsIntervalSeconds             = 14400
	defaultGarbageCollectConfigProjectsEnvironmentsOnInit          = false
	defaultGarbageCollectConfigProjectsEnvironmentsScheduled       = true
	defaultGarbageCollectConfigProjectsEnvironmentsIntervalSeconds = 14400
	defaultGarbageCollectConfigProjectsRefsOnInit                  = false
	defaultGarbageCollectConfigProjectsRefsScheduled               = true
	defaultGarbageCollectConfigProjectsRefsIntervalSeconds         = 1800
	defaultGarbageCollectConfigProjectsRefsMetricsOnInit           = false
	defaultGarbageCollectConfigProjectsRefsMetricsScheduled        = true
	defaultGarbageCollectConfigProjectsRefsMetricsIntervalSeconds  = 300
)

// Config represents what can be defined as a yaml config file
type Config struct {
	// Server related configuration
	Server ServerConfig `yaml:"server"`

	// GitLab related configuration
	Gitlab GitlabConfig `yaml:"gitlab"`

	// Redis related configuration
	Redis RedisConfig `yaml:"redis"`

	// Pull configuration
	Pull PullConfig `yaml:"pull"`

	// GarbageCollect configuration
	GarbageCollect GarbageCollectConfig `yaml:"garbage_collect"`

	// Default parameters which can be overridden at either the Project or Wildcard level
	ProjectDefaults ProjectParameters `yaml:"project_defaults"`

	// List of projects to pull
	Projects []Project `yaml:"projects"`

	// List of wildcards to search projects from
	Wildcards Wildcards `yaml:"wildcards"`
}

// ServerConfig ..
type ServerConfig struct {
	// Enable profiling pages
	EnablePprof bool `yaml:"enable_pprof"`

	// [address:port] to make the process listen upon
	ListenAddress string `yaml:"listen_address"`

	Metrics ServerConfigMetrics `yaml:"metrics"`
	Webhook ServerConfigWebhook `yaml:"webhook"`
}

// ServerConfigMetrics ..
type ServerConfigMetrics struct {
	// Enable /metrics endpoint
	Enabled bool `yaml:"enabled"`

	// Enable OpenMetrics content encoding in prometheus HTTP handler
	EnableOpenmetricsEncoding bool `yaml:"enable_openmetrics_encoding"`
}

// ServerConfigWebhook ..
type ServerConfigWebhook struct {
	// Enable /webhook endpoint to support GitLab requests
	Enabled bool `yaml:"enabled"`

	// Secret token to authenticate legitimate webhook requests coming from the GitLab server
	SecretToken string `yaml:"secret_token"`
}

// GitlabConfig ..
type GitlabConfig struct {
	// The URL of the GitLab server/api
	URL string `yaml:"url"`

	// Token to use to authenticate against the API
	Token string `yaml:"token"`

	// The URL of the GitLab server/api health endpoint (default to /users/sign_in which is publicly available on gitlab.com)
	HealthURL string `yaml:"health_url"`

	// Whether to validate the service is reachable calling HealthURL
	EnableHealthCheck bool `yaml:"enable_health_check"`

	// Whether to skip TLS validation when querying HealthURL
	EnableTLSVerify bool `yaml:"enable_tls_verify"`
}

// RedisConfig ..
type RedisConfig struct {
	// URL used to connect onto the redis endpoint
	// format: redis[s]://[:password@]host[:port][/db-number][?option=value])
	URL string `yaml:"url"`
}

// SchedulerConfig ..
type SchedulerConfig struct {
	OnInit          bool `yaml:"on_init"`
	Scheduled       bool `yaml:"scheduled"`
	IntervalSeconds int  `yaml:"interval_seconds"`
}

// PullConfig ..
type PullConfig struct {
	// Maximum amount of requests per seconds to make against the GitLab API (default: 10)
	MaximumGitLabAPIRequestsPerSecond int `yaml:"maximum_gitlab_api_requests_per_second"`

	// ProjectsFromWildcards configuration
	ProjectsFromWildcards SchedulerConfig `yaml:"projects_from_wildcards"`

	// EnvironmentsFromProjects configuration
	EnvironmentsFromProjects SchedulerConfig `yaml:"environments_from_projects"`

	// RefsFromProjects configuration
	RefsFromProjects SchedulerConfig `yaml:"refs_from_projects"`

	// Metrics configuration
	Metrics SchedulerConfig `yaml:"metrics"`
}

// GarbageCollectConfig ..
type GarbageCollectConfig struct {
	// Projects configuration
	Projects SchedulerConfig `yaml:"projects"`

	// Environments configuration
	Environments SchedulerConfig `yaml:"environments"`

	// Refs configuration
	Refs SchedulerConfig `yaml:"refs"`

	// Metrics configuration
	Metrics SchedulerConfig `yaml:"metrics"`
}

// NewConfig returns a Config with default parameters values
func NewConfig() Config {
	return Config{
		Server: ServerConfig{
			EnablePprof:   defaultServerConfigEnablePprof,
			ListenAddress: defaultServerConfigListenAddress,
			Metrics: ServerConfigMetrics{
				Enabled:                   defaultServerConfigMetricsEnabled,
				EnableOpenmetricsEncoding: defaultServerConfigMetricsEnableOpenmetricsEncoding,
			},
			Webhook: ServerConfigWebhook{
				Enabled: defaultServerConfigWebhookEnabled,
			},
		},
		Gitlab: GitlabConfig{
			URL:               defaultGitlabConfigURL,
			HealthURL:         defaultGitlabConfigHealthURL,
			EnableHealthCheck: defaultGitlabConfigEnableHealthCheck,
			EnableTLSVerify:   defaultGitlabConfigEnableTLSVerify,
		},
		Pull: PullConfig{
			MaximumGitLabAPIRequestsPerSecond: defaultPullConfigMaximumGitLabAPIRequestsPerSecond,
			ProjectsFromWildcards: SchedulerConfig{
				OnInit:          defaultPullConfigProjectsFromWildcardsOnInit,
				Scheduled:       defaultPullConfigProjectsFromWildcardsScheduled,
				IntervalSeconds: defaultPullConfigProjectsFromWildcardsIntervalSeconds,
			},
			EnvironmentsFromProjects: SchedulerConfig{
				OnInit:          defaultPullConfigEnvironmentsFromProjectsOnInit,
				Scheduled:       defaultPullConfigEnvironmentsFromProjectsScheduled,
				IntervalSeconds: defaultPullConfigEnvironmentsFromProjectsIntervalSeconds,
			},
			RefsFromProjects: SchedulerConfig{
				OnInit:          defaultPullConfigRefsFromProjectsOnInit,
				Scheduled:       defaultPullConfigRefsFromProjectsScheduled,
				IntervalSeconds: defaultPullConfigRefsFromProjectsIntervalSeconds,
			},
			Metrics: SchedulerConfig{
				OnInit:          defaultPullConfigMetricsOnInit,
				Scheduled:       defaultPullConfigMetricsScheduled,
				IntervalSeconds: defaultPullConfigMetricsIntervalSeconds,
			},
		},
		GarbageCollect: GarbageCollectConfig{
			Projects: SchedulerConfig{
				OnInit:          defaultGarbageCollectConfigProjectsOnInit,
				Scheduled:       defaultGarbageCollectConfigProjectsScheduled,
				IntervalSeconds: defaultGarbageCollectConfigProjectsIntervalSeconds,
			},
			Environments: SchedulerConfig{
				OnInit:          defaultGarbageCollectConfigProjectsEnvironmentsOnInit,
				Scheduled:       defaultGarbageCollectConfigProjectsEnvironmentsScheduled,
				IntervalSeconds: defaultGarbageCollectConfigProjectsEnvironmentsIntervalSeconds,
			},
			Refs: SchedulerConfig{
				OnInit:          defaultGarbageCollectConfigProjectsRefsOnInit,
				Scheduled:       defaultGarbageCollectConfigProjectsRefsScheduled,
				IntervalSeconds: defaultGarbageCollectConfigProjectsRefsIntervalSeconds,
			},
			Metrics: SchedulerConfig{
				OnInit:          defaultGarbageCollectConfigProjectsRefsMetricsOnInit,
				Scheduled:       defaultGarbageCollectConfigProjectsRefsMetricsScheduled,
				IntervalSeconds: defaultGarbageCollectConfigProjectsRefsMetricsIntervalSeconds,
			},
		},
	}
}

// ParseConfigFile loads a yaml file into a Config structure
func ParseConfigFile(path string) (Config, error) {
	cfg := NewConfig()
	configFile, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return cfg, fmt.Errorf("couldn't open config file : %v", err)
	}

	if err = yaml.Unmarshal(configFile, &cfg); err != nil {
		return cfg, fmt.Errorf("unable to parse config file: %v", err)
	}

	// Hack to fix the missing health endpoint on gitlab.com
	if cfg.Gitlab.URL != "https://gitlab.com" {
		cfg.Gitlab.HealthURL = fmt.Sprintf("%s/-/health", cfg.Gitlab.URL)
	}

	return cfg, nil
}
