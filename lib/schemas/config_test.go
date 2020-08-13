package schemas

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

var cfg = &Config{}

func TestParseInvalidPath(t *testing.T) {
	err := cfg.Parse("/path_do_not_exist")
	assert.Equal(t, fmt.Errorf("couldn't open config file : open /path_do_not_exist: no such file or directory"), err)
}

func TestParseInvalidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("invalid_yaml")
	err = cfg.Parse(f.Name())
	assert.NotNil(t, err)
}

func TestParseEmptyYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("---")
	err = cfg.Parse(f.Name())
	assert.Equal(t, fmt.Errorf("you need to configure at least one project/wildcard to poll, none given"), err)
}

func TestParseValidConfig(t *testing.T) {
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

maximum_gitlab_api_requests_per_second: 1
wildcards_projects_discover_interval_seconds: 2
projects_refs_discover_interval_seconds: 3
projects_refs_polling_interval_seconds: 4
on_init_fetch_refs_from_pipelines: true
on_init_fetch_refs_from_pipelines_depth_limit: 1337
polling_workers: 4

defaults:
  fetch_pipeline_job_metrics: true
  fetch_pipeline_variables: true
  output_sparse_status_metrics: true
  pipeline_variables_regexp: "^CI_"
  refs_regexp: "^dev$"

projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^master|dev$"
  - name: new/project
    refs_regexp: "^main|dev$"

wildcards:
  - owner:
      name: foo
      kind: group
    refs_regexp: "^main|master|1.0$"
    search: 'bar'
    archived: true
`)

	// Reset config var before parsing
	cfg = &Config{}

	err = cfg.Parse(f.Name())
	assert.Nil(t, err)

	expectedCfg := Config{
		Gitlab: struct {
			URL                string "yaml:\"url\""
			Token              string "yaml:\"token\""
			HealthURL          string "yaml:\"health_url\""
			DisableHealthCheck bool   "yaml:\"disable_health_check\""
			DisableTLSVerify   bool   "yaml:\"disable_tls_verify\""
		}{
			URL:                "https://gitlab.example.com",
			HealthURL:          "https://gitlab.example.com/-/health",
			Token:              "xrN14n9-ywvAFxxxxxx",
			DisableHealthCheck: true,
		},
		MaximumGitLabAPIRequestsPerSecond:        1,
		WildcardsProjectsDiscoverIntervalSeconds: 2,
		ProjectsRefsDiscoverIntervalSeconds:      3,
		ProjectsRefsPollingIntervalSeconds:       4,
		OnInitFetchRefsFromPipelines:             true,
		OnInitFetchRefsFromPipelinesDepthLimit:   1337,
		PollingWorkers:                           4,
		Defaults: Parameters{
			FetchPipelineJobMetricsValue:   pointy.Bool(true),
			OutputSparseStatusMetricsValue: pointy.Bool(true),
			FetchPipelineVariablesValue:    pointy.Bool(true),
			PipelineVariablesRegexpValue:   pointy.String("^CI_"),
			RefsRegexpValue:                pointy.String("^dev$"),
		},
		Projects: []Project{
			{
				Name: "foo/project",
				Parameters: Parameters{
					RefsRegexpValue: nil,
				},
			},
			{
				Name: "bar/project",
				Parameters: Parameters{
					RefsRegexpValue: pointy.String("^master|dev$"),
				},
			},
			{
				Name: "new/project",
				Parameters: Parameters{
					RefsRegexpValue: pointy.String("^main|dev$"),
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
				Parameters: Parameters{
					RefsRegexpValue: pointy.String("^main|master|1.0$"),
				},
				Archived: true,
			},
		},
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, *cfg)
}

func TestParseDefaultsValues(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
projects:
  - name: foo/bar
`)

	// Reset config var before parsing
	cfg = &Config{}

	err = cfg.Parse(f.Name())
	assert.Nil(t, err)

	expectedCfg := Config{
		Gitlab: struct {
			URL                string "yaml:\"url\""
			Token              string "yaml:\"token\""
			HealthURL          string "yaml:\"health_url\""
			DisableHealthCheck bool   "yaml:\"disable_health_check\""
			DisableTLSVerify   bool   "yaml:\"disable_tls_verify\""
		}{
			URL:       "https://gitlab.com",
			Token:     "",
			HealthURL: "https://gitlab.com/users/sign_in",
		},
		MaximumGitLabAPIRequestsPerSecond:        defaultMaximumGitLabAPIRequestsPerSecond,
		WildcardsProjectsDiscoverIntervalSeconds: defaultWildcardsProjectsDiscoverIntervalSeconds,
		ProjectsRefsDiscoverIntervalSeconds:      defaultProjectsRefsDiscoverIntervalSeconds,
		ProjectsRefsPollingIntervalSeconds:       defaultProjectsRefsPollingIntervalSeconds,
		DisableOpenmetricsEncoding:               false,
		OnInitFetchRefsFromPipelines:             false,
		OnInitFetchRefsFromPipelinesDepthLimit:   defaultOnInitFetchRefsFromPipelinesDepthLimit,
		PollingWorkers:                           runtime.GOMAXPROCS(0),
		Projects: []Project{
			{
				Name: "foo/bar",
			},
		},
		Wildcards: nil,
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, *cfg)
}

func TestParsePrometheusConfig(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
disable_openmetrics_encoding: true
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^main|master|dev$"
`)

	config := &Config{}
	assert.NoError(t, config.Parse(f.Name()))
	assert.True(t, config.DisableOpenmetricsEncoding)
}

func TestParseConfigWithoutPollingWorkersUsesGOMAXPROCS(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^main|master|dev$"
`)
	config := &Config{}
	assert.NoError(t, config.Parse(f.Name()))
	assert.Equal(t, runtime.GOMAXPROCS(0), config.PollingWorkers)

}

func TestParseConfigHasPipelineVariablesAndDefaultRegex(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
defaults:
  fetch_pipeline_variables: true
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^main|master|dev$"
`)
	cfg := &Config{}
	assert.NoError(t, cfg.Parse(f.Name()))
	assert.NotNil(t, cfg.Defaults.FetchPipelineVariablesValue)
	assert.True(t, *cfg.Defaults.FetchPipelineVariablesValue)
	assert.Len(t, cfg.Projects, 2)
	assert.IsType(t, Project{}, cfg.Projects[0])
	assert.Equal(t, defaultPipelineVariablesRegexp, cfg.Projects[0].PipelineVariablesRegexp(cfg))

	rx := regexp.MustCompile(cfg.Projects[0].PipelineVariablesRegexp(cfg))
	assert.True(t, rx.MatchString("blahblah"))
}
