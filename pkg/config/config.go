package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

// Config represents all the parameters required for the app to be configured properly.
type Config struct {
	// Global ..
	Global Global `yaml:",omitempty"`

	// Log configuration for the exporter
	Log Log `yaml:"log"`

	// OpenTelemetry configuration
	OpenTelemetry OpenTelemetry `yaml:"opentelemetry"`

	// Server related configuration
	Server Server `yaml:"server"`

	// GitLab related configuration
	Gitlab Gitlab `yaml:"gitlab"`

	// Redis related configuration
	Redis Redis `yaml:"redis"`

	// Pull configuration
	Pull Pull `yaml:"pull"`

	// GarbageCollect configuration
	GarbageCollect GarbageCollect `yaml:"garbage_collect"`

	// Default parameters which can be overridden at either the Project or Wildcard level
	ProjectDefaults ProjectParameters `yaml:"project_defaults"`

	// List of projects to pull
	Projects []Project `validate:"unique,at-least-1-project-or-wildcard,dive" yaml:"projects"`

	// List of wildcards to search projects from
	Wildcards []Wildcard `validate:"unique,at-least-1-project-or-wildcard,dive" yaml:"wildcards"`
}

// Log holds runtime logging configuration.
type Log struct {
	// Log level
	Level string `default:"info" validate:"required,oneof=trace debug info warning error fatal panic"`

	// Log format
	Format string `default:"text" validate:"oneof=text json"`
}

// OpenTelemetry related configuration.
type OpenTelemetry struct {
	// gRPC endpoint of the opentelemetry collector
	GRPCEndpoint string `yaml:"grpc_endpoint"`
}

// Server ..
type Server struct {
	// Enable profiling pages
	EnablePprof bool `default:"false" yaml:"enable_pprof"`

	// [address:port] to make the process listen upon
	ListenAddress string `default:":8080" yaml:"listen_address"`

	Metrics ServerMetrics `yaml:"metrics"`
	Webhook ServerWebhook `yaml:"webhook"`
}

// ServerMetrics ..
type ServerMetrics struct {
	// Enable /metrics endpoint
	Enabled bool `default:"true" yaml:"enabled"`

	// Enable OpenMetrics content encoding in prometheus HTTP handler
	EnableOpenmetricsEncoding bool `default:"false" yaml:"enable_openmetrics_encoding"`
}

// ServerWebhook ..
type ServerWebhook struct {
	// Enable /webhook endpoint to support GitLab requests
	Enabled bool `default:"false" yaml:"enabled"`

	// Secret token to authenticate legitimate webhook requests coming from the GitLab server
	SecretToken string `validate:"required_if=Enabled true" yaml:"secret_token"`
}

// Gitlab ..
type Gitlab struct {
	// The URL of the GitLab server/api
	URL string `default:"https://gitlab.com" validate:"required,url" yaml:"url"`

	// Token to use to authenticate against the API
	Token string `validate:"required" yaml:"token"`

	// The URL of the GitLab server/api health endpoint (default to /users/sign_in which is publicly available on gitlab.com)
	HealthURL string `default:"https://gitlab.com/explore" validate:"required,url" yaml:"health_url"`

	// Whether to validate the service is reachable calling HealthURL
	EnableHealthCheck bool `default:"true" yaml:"enable_health_check"`

	// Whether to skip TLS validation when querying HealthURL
	EnableTLSVerify bool `default:"true" yaml:"enable_tls_verify"`

	// Maximum limit for the GitLab API requests/sec
	MaximumRequestsPerSecond int `default:"1" validate:"gte=1" yaml:"maximum_requests_per_second"`

	// Burstable limit for the GitLab API requests/sec
	BurstableRequestsPerSecond int `default:"5" validate:"gte=1" yaml:"burstable_requests_per_second"`

	// Maximum amount of jobs to keep queue, if this limit is reached
	// newly created ones will get dropped. As a best practice you should not change this value.
	// Workarounds to avoid hitting the limit are:
	// - increase polling intervals
	// - increase API rate limit
	// - reduce the amount of projects, refs, environments or metrics you are looking into
	// - leverage webhooks instead of polling schedules
	//
	MaximumJobsQueueSize int `default:"1000" validate:"gte=10" yaml:"maximum_jobs_queue_size"`
}

// Redis ..
type Redis struct {
	// URL used to connect onto the redis endpoint
	// format: redis[s]://[:password@]host[:port][/db-number][?option=value])
	URL string `yaml:"url"`
}

// Pull ..
type Pull struct {
	// ProjectsFromWildcards configuration
	ProjectsFromWildcards struct {
		OnInit          bool `default:"true" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"1800" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"projects_from_wildcards"`

	// EnvironmentsFromProjects configuration
	EnvironmentsFromProjects struct {
		OnInit          bool `default:"true" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"1800" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"environments_from_projects"`

	// RefsFromProjects configuration
	RefsFromProjects struct {
		OnInit          bool `default:"true" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"300" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"refs_from_projects"`

	// Metrics configuration
	Metrics struct {
		OnInit          bool `default:"true" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"30" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"metrics"`
}

// GarbageCollect ..
type GarbageCollect struct {
	// Projects configuration
	Projects struct {
		OnInit          bool `default:"false" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"14400" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"projects"`

	// Environments configuration
	Environments struct {
		OnInit          bool `default:"false" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"14400" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"environments"`

	// Refs configuration
	Refs struct {
		OnInit          bool `default:"false" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"1800" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"refs"`

	// Metrics configuration
	Metrics struct {
		OnInit          bool `default:"false" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"600" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"metrics"`
}

// UnmarshalYAML allows us to correctly hydrate our configuration using some custom logic.
func (c *Config) UnmarshalYAML(v *yaml.Node) (err error) {
	type localConfig struct {
		Log             Log               `yaml:"log"`
		OpenTelemetry   OpenTelemetry     `yaml:"opentelemetry"`
		Server          Server            `yaml:"server"`
		Gitlab          Gitlab            `yaml:"gitlab"`
		Redis           Redis             `yaml:"redis"`
		Pull            Pull              `yaml:"pull"`
		GarbageCollect  GarbageCollect    `yaml:"garbage_collect"`
		ProjectDefaults ProjectParameters `yaml:"project_defaults"`

		Projects  []yaml.Node `yaml:"projects"`
		Wildcards []yaml.Node `yaml:"wildcards"`
	}

	_cfg := localConfig{}
	defaults.MustSet(&_cfg)

	if err = v.Decode(&_cfg); err != nil {
		return
	}

	c.Log = _cfg.Log
	c.OpenTelemetry = _cfg.OpenTelemetry
	c.Server = _cfg.Server
	c.Gitlab = _cfg.Gitlab
	c.Redis = _cfg.Redis
	c.Pull = _cfg.Pull
	c.GarbageCollect = _cfg.GarbageCollect
	c.ProjectDefaults = _cfg.ProjectDefaults

	for _, n := range _cfg.Projects {
		p := c.NewProject()
		if err = n.Decode(&p); err != nil {
			return
		}

		c.Projects = append(c.Projects, p)
	}

	for _, n := range _cfg.Wildcards {
		w := c.NewWildcard()
		if err = n.Decode(&w); err != nil {
			return
		}

		c.Wildcards = append(c.Wildcards, w)
	}

	return
}

// ToYAML ..
func (c Config) ToYAML() string {
	c.Global = Global{}
	c.Server.Webhook.SecretToken = "*******"
	c.Gitlab.Token = "*******"

	b, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// Validate will throw an error if the Config parameters are whether incomplete or incorrects.
func (c Config) Validate() error {
	if validate == nil {
		validate = validator.New()
		_ = validate.RegisterValidation("at-least-1-project-or-wildcard", ValidateAtLeastOneProjectOrWildcard)
	}

	return validate.Struct(c)
}

// SchedulerConfig ..
type SchedulerConfig struct {
	OnInit          bool
	Scheduled       bool
	IntervalSeconds int
}

// Log returns some logging fields to showcase the configuration to the enduser.
func (sc SchedulerConfig) Log() log.Fields {
	onInit, scheduled := "no", "no"
	if sc.OnInit {
		onInit = "yes"
	}

	if sc.Scheduled {
		scheduled = fmt.Sprintf("every %vs", sc.IntervalSeconds)
	}

	return log.Fields{
		"on-init":   onInit,
		"scheduled": scheduled,
	}
}

// ValidateAtLeastOneProjectOrWildcard implements validator.Func
// assess that we have at least one projet or wildcard configured.
func ValidateAtLeastOneProjectOrWildcard(v validator.FieldLevel) bool {
	return v.Parent().FieldByName("Projects").Len() > 0 || v.Parent().FieldByName("Wildcards").Len() > 0
}

// New returns a new config with the default parameters.
func New() (c Config) {
	defaults.MustSet(&c)

	return
}

// NewProject returns a new project with the config default parameters.
func (c Config) NewProject() (p Project) {
	p.ProjectParameters = c.ProjectDefaults

	return
}

// NewWildcard returns a new wildcard with the config default parameters.
func (c Config) NewWildcard() (w Wildcard) {
	w.ProjectParameters = c.ProjectDefaults

	return
}
