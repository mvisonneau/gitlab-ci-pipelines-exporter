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
	assert.Equal(t, fmt.Errorf("couldn't open config file : open /path_do_not_exist: no such file or directory"), ParseConfig("/path_do_not_exist", &Config{}))
}

func TestParseConfigInvalidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("invalid_yaml")
	assert.Error(t, ParseConfig(f.Name(), &Config{}))
}

func TestParseConfigEmptyYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("---")
	assert.Equal(t, fmt.Errorf("you need to configure at least one project/wildcard to poll, none given"), ParseConfig(f.Name(), &Config{}))
}

func TestParseConfigValidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx
  health_url: https://gitlab.example.com/-/health
  disable_health_check: true
  disable_tls_verify: true

disable_openmetrics_encoding: true

pull:
  maximum_gitlab_api_requests_per_second: 1
  metrics:
    on_init: false
    scheduled: false
    interval_seconds: 1
  projects_from_wildcards:
    on_init: false
    scheduled: false
    interval_seconds: 2
  project_refs_from_branches_tags_and_mrs:
    on_init: false
    scheduled: false
    interval_seconds: 3

project_defaults:
  output_sparse_status_metrics: false
  pull:
    refs:
      regexp: "^baz$"
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
      refs:
        regexp: "^foo$"
  - name: new/project
    pull:
      refs:
        regexp: "^bar$"

wildcards:
  - owner:
      name: foo
      kind: group
    search: 'bar'
    archived: true
    pull:
      refs:
        regexp: "^yolo$"
`)

	cfg := &Config{}
	assert.NoError(t, ParseConfig(f.Name(), cfg))

	expectedCfg := Config{
		Gitlab: GitlabConfig{
			URL:                "https://gitlab.example.com",
			HealthURL:          "https://gitlab.example.com/-/health",
			Token:              "xrN14n9-ywvAFxxxxxx",
			DisableHealthCheck: true,
			DisableTLSVerify:   true,
		},
		DisableOpenmetricsEncoding: true,
		Pull: PullConfig{
			MaximumGitLabAPIRequestsPerSecondValue: pointy.Int(1),
			Metrics: PullConfigMetrics{
				OnInitValue:          pointy.Bool(false),
				ScheduledValue:       pointy.Bool(false),
				IntervalSecondsValue: pointy.Int(1),
			},
			ProjectsFromWildcards: PullConfigProjectsFromWildcards{
				OnInitValue:          pointy.Bool(false),
				ScheduledValue:       pointy.Bool(false),
				IntervalSecondsValue: pointy.Int(2),
			},
			ProjectRefsFromBranchesTagsMergeRequests: PullConfigProjectRefsFromBranchesTagsMergeRequests{
				OnInitValue:          pointy.Bool(false),
				ScheduledValue:       pointy.Bool(false),
				IntervalSecondsValue: pointy.Int(3),
			},
		},
		ProjectDefaults: ProjectParameters{
			OutputSparseStatusMetricsValue: pointy.Bool(false),
			Pull: ProjectPull{
				Refs: ProjectPullRefs{
					RegexpValue: pointy.String("^baz$"),
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
						Refs: ProjectPullRefs{
							RegexpValue: pointy.String("^foo$"),
						},
					},
				},
			},
			{
				Name: "new/project",
				ProjectParameters: ProjectParameters{
					Pull: ProjectPull{
						Refs: ProjectPullRefs{
							RegexpValue: pointy.String("^bar$"),
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
						Refs: ProjectPullRefs{
							RegexpValue: pointy.String("^yolo$"),
						},
					},
				},
				Archived: true,
			},
		},
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, *cfg)
}

func TestParseConfigDefaultsValues(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
wildcards:
  - {}
`)

	cfg := &Config{}
	assert.NoError(t, ParseConfig(f.Name(), cfg))

	expectedCfg := Config{
		Gitlab: GitlabConfig{
			URL:       "https://gitlab.com",
			Token:     "",
			HealthURL: "https://gitlab.com/explore",
		},
		DisableOpenmetricsEncoding: false,
		Wildcards: Wildcards{
			{},
		},
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, *cfg)

	// Validate default values
	assert.Equal(t, defaultPullConfigMaximumGitLabAPIRequestsPerSecond, cfg.Pull.MaximumGitLabAPIRequestsPerSecond())

	assert.Equal(t, defaultPullConfigMetricsOnInit, cfg.Pull.Metrics.OnInit())
	assert.Equal(t, defaultPullConfigMetricsScheduled, cfg.Pull.Metrics.Scheduled())
	assert.Equal(t, defaultPullConfigMetricsIntervalSeconds, cfg.Pull.Metrics.IntervalSeconds())

	assert.Equal(t, defaultPullConfigProjectsFromWildcardsOnInit, cfg.Pull.ProjectsFromWildcards.OnInit())
	assert.Equal(t, defaultPullConfigProjectsFromWildcardsScheduled, cfg.Pull.ProjectsFromWildcards.Scheduled())
	assert.Equal(t, defaultPullConfigProjectsFromWildcardsIntervalSeconds, cfg.Pull.ProjectsFromWildcards.IntervalSeconds())

	assert.Equal(t, defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsOnInit, cfg.Pull.ProjectRefsFromBranchesTagsMergeRequests.OnInit())
	assert.Equal(t, defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsScheduled, cfg.Pull.ProjectRefsFromBranchesTagsMergeRequests.Scheduled())
	assert.Equal(t, defaultPullConfigProjectRefsFromBranchesTagsMergeRequestsIntervalSeconds, cfg.Pull.ProjectRefsFromBranchesTagsMergeRequests.IntervalSeconds())

	assert.Equal(t, defaultProjectOutputSparseStatusMetrics, cfg.ProjectDefaults.OutputSparseStatusMetrics())
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

wildcards:
  - {}
`)

	cfg := &Config{}
	assert.NoError(t, ParseConfig(f.Name(), cfg))
	assert.Equal(t, "https://gitlab.example.com/-/health", cfg.Gitlab.HealthURL)
}
