package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFileInvalidPath(t *testing.T) {
	cfg, err := ParseFile("/path_do_not_exist.yml")
	assert.Error(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestParseInvalidYaml(t *testing.T) {
	cfg, err := Parse(FormatYAML, []byte("invalid_yaml"))
	assert.Error(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestParseValidYaml(t *testing.T) {
	yamlConfig := `
---
log:
  level: trace
  format: json

opentelemetry:
  grpc_endpoint: otlp-collector:4317

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
  maximum_requests_per_second: 2

redis:
  url: redis://popopo:1337

pull:
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
      regexp: "^baz$"
    refs:
      branches:
        enabled: false
        regexp: "^foo$"
        most_recent: 1
        max_age_seconds: 1
        exclude_deleted: false
      tags:
        enabled: false
        regexp: "^bar$"
        most_recent: 2
        max_age_seconds: 2
        exclude_deleted: false
      merge_requests:
        enabled: true
        most_recent: 3
        max_age_seconds: 3
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
        regexp: "^foo$"
      refs:
        branches:
          regexp: "^foo$"
          max_age_seconds: 2

wildcards:
  - owner:
      name: foo
      kind: group
    search: 'bar'
    archived: true
    pull:
      environments:
        enabled: false
        regexp: "^foo$"
      refs:
        branches:
          regexp: "^yolo$"
          max_age_seconds: 4
`

	cfg, err := Parse(FormatYAML, []byte(yamlConfig))
	assert.NoError(t, err)

	xcfg := New()

	xcfg.Log.Level = "trace"
	xcfg.Log.Format = "json"

	xcfg.OpenTelemetry.GRPCEndpoint = "otlp-collector:4317"

	xcfg.Server.EnablePprof = true
	xcfg.Server.ListenAddress = ":1025"
	xcfg.Server.Metrics.Enabled = false
	xcfg.Server.Metrics.EnableOpenmetricsEncoding = false

	xcfg.Server.Webhook.Enabled = true
	xcfg.Server.Webhook.SecretToken = "secret"

	xcfg.Gitlab.URL = "https://gitlab.example.com"
	xcfg.Gitlab.HealthURL = "https://gitlab.example.com/-/health"
	xcfg.Gitlab.Token = "xrN14n9-ywvAFxxxxxx"
	xcfg.Gitlab.EnableHealthCheck = false
	xcfg.Gitlab.EnableTLSVerify = false
	xcfg.Gitlab.MaximumRequestsPerSecond = 2

	xcfg.Redis.URL = "redis://popopo:1337"

	xcfg.Pull.ProjectsFromWildcards.OnInit = false
	xcfg.Pull.ProjectsFromWildcards.Scheduled = false
	xcfg.Pull.ProjectsFromWildcards.IntervalSeconds = 1

	xcfg.Pull.EnvironmentsFromProjects.OnInit = false
	xcfg.Pull.EnvironmentsFromProjects.Scheduled = false
	xcfg.Pull.EnvironmentsFromProjects.IntervalSeconds = 2

	xcfg.Pull.RefsFromProjects.OnInit = false
	xcfg.Pull.RefsFromProjects.Scheduled = false
	xcfg.Pull.RefsFromProjects.IntervalSeconds = 3

	xcfg.Pull.Metrics.OnInit = false
	xcfg.Pull.Metrics.Scheduled = false
	xcfg.Pull.Metrics.IntervalSeconds = 4

	xcfg.GarbageCollect.Projects.OnInit = true
	xcfg.GarbageCollect.Projects.Scheduled = false
	xcfg.GarbageCollect.Projects.IntervalSeconds = 1

	xcfg.GarbageCollect.Environments.OnInit = true
	xcfg.GarbageCollect.Environments.Scheduled = false
	xcfg.GarbageCollect.Environments.IntervalSeconds = 2

	xcfg.GarbageCollect.Refs.OnInit = true
	xcfg.GarbageCollect.Refs.Scheduled = false
	xcfg.GarbageCollect.Refs.IntervalSeconds = 3

	xcfg.GarbageCollect.Metrics.OnInit = true
	xcfg.GarbageCollect.Metrics.Scheduled = false
	xcfg.GarbageCollect.Metrics.IntervalSeconds = 4

	xcfg.ProjectDefaults.OutputSparseStatusMetrics = false

	xcfg.ProjectDefaults.Pull.Environments.Enabled = true
	xcfg.ProjectDefaults.Pull.Environments.Regexp = `^baz$`

	xcfg.ProjectDefaults.Pull.Refs.Branches.Enabled = false
	xcfg.ProjectDefaults.Pull.Refs.Branches.Regexp = `^foo$`
	xcfg.ProjectDefaults.Pull.Refs.Branches.MostRecent = 1
	xcfg.ProjectDefaults.Pull.Refs.Branches.MaxAgeSeconds = 1
	xcfg.ProjectDefaults.Pull.Refs.Branches.ExcludeDeleted = false

	xcfg.ProjectDefaults.Pull.Refs.Tags.Enabled = false
	xcfg.ProjectDefaults.Pull.Refs.Tags.Regexp = `^bar$`
	xcfg.ProjectDefaults.Pull.Refs.Tags.MostRecent = 2
	xcfg.ProjectDefaults.Pull.Refs.Tags.MaxAgeSeconds = 2
	xcfg.ProjectDefaults.Pull.Refs.Tags.ExcludeDeleted = false

	xcfg.ProjectDefaults.Pull.Refs.MergeRequests.Enabled = true
	xcfg.ProjectDefaults.Pull.Refs.MergeRequests.MostRecent = 3
	xcfg.ProjectDefaults.Pull.Refs.MergeRequests.MaxAgeSeconds = 3

	xcfg.ProjectDefaults.Pull.Pipeline.Jobs.Enabled = true
	xcfg.ProjectDefaults.Pull.Pipeline.Variables.Enabled = true
	xcfg.ProjectDefaults.Pull.Pipeline.Variables.Regexp = `^CI_`

	p1 := NewProject("foo/project")
	p1.ProjectParameters = xcfg.ProjectDefaults

	p2 := NewProject("bar/project")
	p2.ProjectParameters = xcfg.ProjectDefaults

	p2.Pull.Environments.Enabled = false
	p2.Pull.Environments.Regexp = `^foo$`
	p2.Pull.Refs.Branches.Regexp = `^foo$`
	p2.Pull.Refs.Branches.MaxAgeSeconds = 2

	xcfg.Projects = []Project{p1, p2}

	w1 := NewWildcard()
	w1.ProjectParameters = xcfg.ProjectDefaults
	w1.Search = "bar"
	w1.Archived = true
	w1.Owner.Name = "foo"
	w1.Owner.Kind = "group"
	w1.Pull.Environments.Enabled = false
	w1.Pull.Environments.Regexp = `^foo$`
	w1.Pull.Refs.Branches.Regexp = `^yolo$`
	w1.Pull.Refs.Branches.MaxAgeSeconds = 4

	xcfg.Wildcards = []Wildcard{w1}

	// Test variable assignments
	assert.Equal(t, xcfg, cfg)
}

func TestParseConfigSelfHostedGitLab(t *testing.T) {
	yamlConfig := `
---
gitlab:
  url: https://gitlab.example.com
`
	cfg, err := Parse(
		FormatYAML,
		[]byte(yamlConfig),
	)

	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com/-/health", cfg.Gitlab.HealthURL)
}
