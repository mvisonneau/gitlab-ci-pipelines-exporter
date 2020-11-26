package schemas

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func TestParseConfigInvalidPath(t *testing.T) {
	cfg, err := ParseConfigFile("/path_do_not_exist")
	assert.Equal(t, fmt.Errorf("couldn't open config file : open /path_do_not_exist: no such file or directory"), err)
	assert.Equal(t, NewConfig(), cfg)
}

func TestParseConfigFileInvalidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("invalid_yaml")
	cfg, err := ParseConfigFile(f.Name())
	assert.Error(t, err)
	assert.Equal(t, NewConfig(), cfg)
}

func TestParseConfigValidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
server:
  enable_pprof: true
  listen_address: :1025

  metrics:
    enabled: false
    enable_openmetrics_encoding: false

  webhook:
    enabled: true
    secret_token: secret

gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx
  health_url: https://gitlab.example.com/-/health
  enable_health_check: false
  enable_tls_verify: false

redis:
  url: redis://popopo:1337

pull:
  maximum_gitlab_api_requests_per_second: 1
  projects_from_wildcards:
    on_init: false
    scheduled: false
    interval_seconds: 1
  environments_from_projects:
    on_init: false
    scheduled: false
    interval_seconds: 2
  refs_from_projects:
    on_init: false
    scheduled: false
    interval_seconds: 3
  metrics:
    on_init: false
    scheduled: false
    interval_seconds: 4

garbage_collect:
  projects:
    on_init: true
    scheduled: false
    interval_seconds: 1
  environments:
    on_init: true
    scheduled: false
    interval_seconds: 2
  refs:
    on_init: true
    scheduled: false
    interval_seconds: 3
  metrics:
    on_init: true
    scheduled: false
    interval_seconds: 4

project_defaults:
  output_sparse_status_metrics: false
  pull:
    environments:
      enabled: true
      name_regexp: "^baz$"
      tags_regexp: "^blah$"
    refs:
      regexp: "^baz$"
      max_age_seconds: 1
      from:
        pipelines:
          enabled: true
          depth: 1
        merge_requests:
          enabled: true
          depth: 2
    pipeline:
      jobs:
        enabled: true
      variables:
        enabled: true
        regexp: "^CI_"

projects:
  - name: foo/project
  - name: bar/project
    pull:
      environments:
        enabled: false
        name_regexp: "^foo$"
        tags_regexp: "^foo$"
      refs:
        regexp: "^foo$"
        max_age_seconds: 2
  - name: new/project
    pull:
      environments:
        enabled: false
        name_regexp: "^foo$"
        tags_regexp: "^foo$"
      refs:
        regexp: "^bar$"
        max_age_seconds: 3

wildcards:
  - owner:
      name: foo
      kind: group
    search: 'bar'
    archived: true
    pull:
      environments:
        enabled: false
        name_regexp: "^foo$"
        tags_regexp: "^foo$"
      refs:
        regexp: "^yolo$"
        max_age_seconds: 4
`)

	cfg, err := ParseConfigFile(f.Name())
	assert.NoError(t, err)

	expectedCfg := Config{
		Server: ServerConfig{
			EnablePprof:   true,
			ListenAddress: ":1025",
			Metrics: ServerConfigMetrics{
				Enabled:                   false,
				EnableOpenmetricsEncoding: false,
			},
			Webhook: ServerConfigWebhook{
				Enabled:     true,
				SecretToken: "secret",
			},
		},
		Gitlab: GitlabConfig{
			URL:               "https://gitlab.example.com",
			HealthURL:         "https://gitlab.example.com/-/health",
			Token:             "xrN14n9-ywvAFxxxxxx",
			EnableHealthCheck: false,
			EnableTLSVerify:   false,
		},
		Redis: RedisConfig{
			URL: "redis://popopo:1337",
		},
		Pull: PullConfig{
			MaximumGitLabAPIRequestsPerSecond: 1,
			ProjectsFromWildcards: SchedulerConfig{
				OnInit:          false,
				Scheduled:       false,
				IntervalSeconds: 1,
			},
			EnvironmentsFromProjects: SchedulerConfig{
				OnInit:          false,
				Scheduled:       false,
				IntervalSeconds: 2,
			},
			RefsFromProjects: SchedulerConfig{
				OnInit:          false,
				Scheduled:       false,
				IntervalSeconds: 3,
			},
			Metrics: SchedulerConfig{
				OnInit:          false,
				Scheduled:       false,
				IntervalSeconds: 4,
			},
		},
		GarbageCollect: GarbageCollectConfig{
			Projects: SchedulerConfig{
				OnInit:          true,
				Scheduled:       false,
				IntervalSeconds: 1,
			},
			Environments: SchedulerConfig{
				OnInit:          true,
				Scheduled:       false,
				IntervalSeconds: 2,
			},
			Refs: SchedulerConfig{
				OnInit:          true,
				Scheduled:       false,
				IntervalSeconds: 3,
			},
			Metrics: SchedulerConfig{
				OnInit:          true,
				Scheduled:       false,
				IntervalSeconds: 4,
			},
		},
		ProjectDefaults: ProjectParameters{
			OutputSparseStatusMetricsValue: pointy.Bool(false),
			Pull: ProjectPull{
				Environments: ProjectPullEnvironments{
					EnabledValue:    pointy.Bool(true),
					NameRegexpValue: pointy.String("^baz$"),
					TagsRegexpValue: pointy.String("^blah$"),
				},
				Refs: ProjectPullRefs{
					RegexpValue:        pointy.String("^baz$"),
					MaxAgeSecondsValue: pointy.Uint(1),
					From: ProjectPullRefsFrom{
						Pipelines: ProjectPullRefsFromPipelines{
							EnabledValue: pointy.Bool(true),
							DepthValue:   pointy.Int(1),
						},
						MergeRequests: ProjectPullRefsFromMergeRequests{
							EnabledValue: pointy.Bool(true),
							DepthValue:   pointy.Int(2),
						},
					},
				},
				Pipeline: ProjectPullPipeline{
					Jobs: ProjectPullPipelineJobs{
						EnabledValue: pointy.Bool(true),
					},
					Variables: ProjectPullPipelineVariables{
						EnabledValue: pointy.Bool(true),
						RegexpValue:  pointy.String("^CI_"),
					},
				},
			},
		},
		Projects: []Project{
			{
				Name: "foo/project",
			},
			{
				Name: "bar/project",
				ProjectParameters: ProjectParameters{
					Pull: ProjectPull{
						Environments: ProjectPullEnvironments{
							EnabledValue:    pointy.Bool(false),
							NameRegexpValue: pointy.String("^foo$"),
							TagsRegexpValue: pointy.String("^foo$"),
						},
						Refs: ProjectPullRefs{
							RegexpValue:        pointy.String("^foo$"),
							MaxAgeSecondsValue: pointy.Uint(2),
						},
					},
				},
			},
			{
				Name: "new/project",
				ProjectParameters: ProjectParameters{
					Pull: ProjectPull{
						Environments: ProjectPullEnvironments{
							EnabledValue:    pointy.Bool(false),
							NameRegexpValue: pointy.String("^foo$"),
							TagsRegexpValue: pointy.String("^foo$"),
						},
						Refs: ProjectPullRefs{
							RegexpValue:        pointy.String("^bar$"),
							MaxAgeSecondsValue: pointy.Uint(3),
						},
					},
				},
			},
		},
		Wildcards: []Wildcard{
			{
				Search: "bar",
				Owner: struct {
					Name             string `yaml:"name"`
					Kind             string `yaml:"kind"`
					IncludeSubgroups bool   `yaml:"include_subgroups"`
				}{
					Name: "foo",
					Kind: "group",
				},
				ProjectParameters: ProjectParameters{
					Pull: ProjectPull{
						Environments: ProjectPullEnvironments{
							EnabledValue:    pointy.Bool(false),
							NameRegexpValue: pointy.String("^foo$"),
							TagsRegexpValue: pointy.String("^foo$"),
						},
						Refs: ProjectPullRefs{
							RegexpValue:        pointy.String("^yolo$"),
							MaxAgeSecondsValue: pointy.Uint(4),
						},
					},
				},
				Archived: true,
			},
		},
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, cfg)
}

func TestParseConfigDefaultsValues(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
`)

	cfg, err := ParseConfigFile(f.Name())
	assert.NoError(t, err)
	expectedCfg := NewConfig()

	// Test variable assignments
	assert.Equal(t, expectedCfg, cfg)

	// Validate project default values
	assert.Equal(t, defaultProjectOutputSparseStatusMetrics, cfg.ProjectDefaults.OutputSparseStatusMetrics())

	assert.Equal(t, defaultProjectPullEnvironmentsEnabled, cfg.ProjectDefaults.Pull.Environments.Enabled())
	assert.Equal(t, defaultProjectPullEnvironmentsNameRegexp, cfg.ProjectDefaults.Pull.Environments.NameRegexp())
	assert.Equal(t, defaultProjectPullEnvironmentsTagsRegexp, cfg.ProjectDefaults.Pull.Environments.TagsRegexp())

	assert.Equal(t, defaultProjectPullRefsRegexp, cfg.ProjectDefaults.Pull.Refs.Regexp())
	assert.Equal(t, defaultProjectPullRefsFromPipelinesEnabled, cfg.ProjectDefaults.Pull.Refs.From.Pipelines.Enabled())
	assert.Equal(t, defaultProjectPullRefsFromPipelinesDepth, cfg.ProjectDefaults.Pull.Refs.From.Pipelines.Depth())

	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsEnabled, cfg.ProjectDefaults.Pull.Refs.From.MergeRequests.Enabled())
	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsDepth, cfg.ProjectDefaults.Pull.Refs.From.MergeRequests.Depth())

	assert.Equal(t, defaultProjectPullPipelineJobsEnabled, cfg.ProjectDefaults.Pull.Pipeline.Jobs.Enabled())

	assert.Equal(t, defaultProjectPullPipelineVariablesEnabled, cfg.ProjectDefaults.Pull.Pipeline.Variables.Enabled())
	assert.Equal(t, defaultProjectPullPipelineVariablesRegexp, cfg.ProjectDefaults.Pull.Pipeline.Variables.Regexp())
}

func TestParseConfigSelfHostedGitLab(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
gitlab:
  url: https://gitlab.example.com
`)

	cfg, err := ParseConfigFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/-/health", cfg.Gitlab.HealthURL)
}
