package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

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
projects_polling_interval_seconds: 2
refs_polling_interval_seconds: 3
pipelines_polling_interval_seconds: 4
on_init_fetch_refs_from_pipelines: true
on_init_fetch_refs_from_pipelines_depth_limit: 1337
default_refs_regexp: "^dev$"
maximum_projects_poller_workers: 4

projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^master|dev$"

wildcards:
  - owner:
      name: foo
      kind: group
    refs_regexp: "^master|1.0$"
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
		MaximumGitLabAPIRequestsPerSecond:      1,
		ProjectsPollingIntervalSeconds:         2,
		RefsPollingIntervalSeconds:             3,
		PipelinesPollingIntervalSeconds:        4,
		OnInitFetchRefsFromPipelines:           true,
		OnInitFetchRefsFromPipelinesDepthLimit: 1337,
		DefaultRefsRegexp:                      "^dev$",
		MaximumProjectsPollingWorkers:          4,
		PipelineVariablesFilterRegexp:          variablesCatchallRegex,
		Projects: []Project{
			{
				Name:       "foo/project",
				RefsRegexp: "",
			},
			{
				Name:       "bar/project",
				RefsRegexp: "^master|dev$",
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
				RefsRegexp: "^master|1.0$",
				Archived:   true,
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
		MaximumGitLabAPIRequestsPerSecond:      defaultMaximumGitLabAPIRequestsPerSecond,
		ProjectsPollingIntervalSeconds:         defaultProjectsPollingIntervalSeconds,
		RefsPollingIntervalSeconds:             defaultRefsPollingIntervalSeconds,
		PipelinesPollingIntervalSeconds:        defaultPipelinesPollingIntervalSeconds,
		DisableOpenmetricsEncoding:             false,
		OnInitFetchRefsFromPipelines:           false,
		OnInitFetchRefsFromPipelinesDepthLimit: defaultOnInitFetchRefsFromPipelinesDepthLimit,
		DefaultRefsRegexp:                      "",
		MaximumProjectsPollingWorkers:          runtime.GOMAXPROCS(0),
		FetchPipelineVariables:                 false,
		PipelineVariablesFilterRegexp:          variablesCatchallRegex,
		Projects: []Project{
			{
				Name:       "foo/bar",
				RefsRegexp: "",
			},
		},
		Wildcards: nil,
	}

	// Test variable assignments
	assert.Equal(t, expectedCfg, *cfg)
}

func TestMergeWithContext(t *testing.T) {
	expectedFileToken := "file-foo-bar"
	expectedCtxToken := "ctx-foo-bar"

	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
gitlab:
  token: file-foo-bar
projects:
  - name: foo/bar
`)

	// Reset config var before parsing
	cfg = &Config{}

	err = cfg.Parse(f.Name())
	assert.Nil(t, err)
	assert.Equal(t, expectedFileToken, cfg.Gitlab.Token)

	set := flag.NewFlagSet("", 0)
	set.String("gitlab-token", expectedCtxToken, "")

	ctx := cli.NewContext(nil, set, nil)
	cfg.MergeWithContext(ctx)

	assert.Equal(t, expectedCtxToken, cfg.Gitlab.Token)
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
    refs_regexp: "^master|dev$"
`)

	config := &Config{}
	assert.NoError(t, config.Parse(f.Name()))
	assert.True(t, config.DisableOpenmetricsEncoding)
}

func TestParseConfigWithoutProjectWorkersUsesGOMAXPROCS(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^master|dev$"
`)
	config := &Config{}
	assert.NoError(t, config.Parse(f.Name()))
	assert.Equal(t, runtime.GOMAXPROCS(0), config.MaximumProjectsPollingWorkers)

}

func TestParseConfigHasPipelineVariablesAndDefaultRegex(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
fetch_pipeline_variables: true
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^master|dev$"	
`)
	config := &Config{}
	assert.NoError(t, config.Parse(f.Name()))
	assert.True(t, config.FetchPipelineVariables)
	assert.Equal(t, "\\.*", config.PipelineVariablesFilterRegexp)

	rx := regexp.MustCompile(config.PipelineVariablesFilterRegexp)
	assert.True(t, rx.MatchString("blahblah"))
}
