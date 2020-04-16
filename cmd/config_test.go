package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestParseInvalidPath(t *testing.T) {
	err := cfg.Parse("/path_do_not_exist")
	assert.Equal(t, fmt.Errorf("Couldn't open config file : open /path_do_not_exist: no such file or directory"), err)
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
	assert.Equal(t, fmt.Errorf("You need to configure at least one project/wildcard to poll, none given"), err)
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
  skip_tls_verify: true

maximum_gitlab_api_requests_per_second: 1
projects_polling_interval_seconds: 2
refs_polling_interval_seconds: 3
pipelines_polling_interval_seconds: 4
on_init_fetch_refs_from_pipelines: true
on_init_fetch_refs_from_pipelines_depth_limit: 1337
default_refs: "^dev$"

projects:
  - name: foo/project
  - name: bar/project
    refs: "^master|dev$"

wildcards:
  - owner:
      name: foo
      kind: group
    refs: "^master|1.0$"
    search: 'bar'
    archived: true
`)

	// Reset config var before parsing
	cfg = &Config{}

	err = cfg.Parse(f.Name())
	assert.Nil(t, err)

	expectedCfg := Config{
		Gitlab: struct {
			URL           string "yaml:\"url\""
			Token         string "yaml:\"token\""
			HealthURL     string "yaml:\"health_url\""
			SkipTLSVerify bool   "yaml:\"skip_tls_verify\""
		}{
			URL:           "https://gitlab.example.com",
			HealthURL:     "https://gitlab.example.com/-/health",
			Token:         "xrN14n9-ywvAFxxxxxx",
			SkipTLSVerify: true,
		},
		MaximumGitLabAPIRequestsPerSecond:      1,
		ProjectsPollingIntervalSeconds:         2,
		RefsPollingIntervalSeconds:             3,
		PipelinesPollingIntervalSeconds:        4,
		OnInitFetchRefsFromPipelines:           true,
		OnInitFetchRefsFromPipelinesDepthLimit: 1337,
		DefaultRefsRegexp:                      "^dev$",
		Projects: []Project{
			{
				Name: "foo/project",
				Refs: "",
			},
			{
				Name: "bar/project",
				Refs: "^master|dev$",
			},
		},
		Wildcards: []Wildcard{
			{
				Search: "bar",
				Owner: struct {
					Name             string
					Kind             string
					IncludeSubgroups bool `yaml:"include_subgroups"`
				}{
					Name: "foo",
					Kind: "group",
				},
				Refs:     "^master|1.0$",
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
			URL           string "yaml:\"url\""
			Token         string "yaml:\"token\""
			HealthURL     string "yaml:\"health_url\""
			SkipTLSVerify bool   "yaml:\"skip_tls_verify\""
		}{
			URL:           "https://gitlab.com",
			Token:         "",
			HealthURL:     "https://gitlab.com/users/sign_in",
			SkipTLSVerify: false,
		},
		MaximumGitLabAPIRequestsPerSecond:      defaultMaximumGitLabAPIRequestsPerSecond,
		ProjectsPollingIntervalSeconds:         defaultProjectsPollingIntervalSeconds,
		RefsPollingIntervalSeconds:             defaultRefsPollingIntervalSeconds,
		PipelinesPollingIntervalSeconds:        defaultPipelinesPollingIntervalSeconds,
		OnInitFetchRefsFromPipelines:           false,
		OnInitFetchRefsFromPipelinesDepthLimit: defaultOnInitFetchRefsFromPipelinesDepthLimit,
		DefaultRefsRegexp:                      "",
		Projects: []Project{
			{
				Name: "foo/bar",
				Refs: "",
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
