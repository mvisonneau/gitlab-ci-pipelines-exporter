package cmd

import (
	"io/ioutil"
	"os"

	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseInvalidPath(t *testing.T) {
	err := cfg.Parse("/path_do_not_exist")
	if err == nil {
		t.Fatal("Expected config parser to return an error")
	}

	if err.Error() != "Couldn't open config file : open /path_do_not_exist: no such file or directory" {
		t.Fatalf("Unexpected returned error : %s", err.Error())
	}
}

func TestParseInvalidYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	if err != nil {
		t.Fatal("Could not create temporary test files")
	}
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("invalid_yaml")
	err = cfg.Parse(f.Name())
	if err == nil {
		t.Fatal("Expected config parser to return an error")
	}
}

func TestParseEmptyYaml(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	if err != nil {
		t.Fatal("Could not create temporary test files")
	}
	defer os.Remove(f.Name())

	// Invalid YAML content
	f.WriteString("---")
	err = cfg.Parse(f.Name())
	if err == nil {
		t.Fatal("Expected config parser to return an error")
	}

	if err.Error() != "You need to configure at least one project/wildcard to poll, none given" {
		t.Fatalf("Unexpected returned error : %s", err.Error())
	}
}

func TestParseValidConfig(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	if err != nil {
		t.Fatal("Could not create temporary test files")
	}
	defer os.Remove(f.Name())

	// Valid minimal configuration
	f.WriteString(`
---
gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx
  health_url: https://gitlab.example.com/-/health
  skip_tls_verify: true

projects_polling_interval_seconds: 1
refs_polling_interval_seconds: 2
pipelines_polling_interval_seconds: 3
pipelines_max_polling_interval_seconds: 4
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
`)

	// Reset config var before parsing
	cfg = &Config{}

	err = cfg.Parse(f.Name())
	if err != nil {
		t.Fatalf("Did not expect an error, got %s", err.Error())
	}

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
		ProjectsPollingIntervalSeconds:     1,
		RefsPollingIntervalSeconds:         2,
		PipelinesPollingIntervalSeconds:    3,
		PipelinesMaxPollingIntervalSeconds: 4,
		DefaultRefsRegexp:                  "^dev$",
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
				Refs: "^master|1.0$",
			},
		},
	}

	// Test variable assignments
	if !cmp.Equal(*cfg, expectedCfg) {
		t.Fatalf("Diff of expected/got config :\n %v", cmp.Diff(*cfg, expectedCfg))
	}
}

func TestParseDefaultsValues(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "test-")
	if err != nil {
		t.Fatal("Could not create temporary test files")
	}
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
	if err != nil {
		t.Fatalf("Did not expect an error, got %s", err.Error())
	}

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
		ProjectsPollingIntervalSeconds:     defaultProjectsPollingIntervalSeconds,
		RefsPollingIntervalSeconds:         defaultRefsPollingIntervalSeconds,
		PipelinesPollingIntervalSeconds:    defaultPipelinesPollingIntervalSeconds,
		PipelinesMaxPollingIntervalSeconds: defaultPipelinesMaxPollingIntervalSeconds,
		DefaultRefsRegexp:                  "",
		Projects: []Project{
			{
				Name: "foo/bar",
				Refs: "",
			},
		},
		Wildcards: nil,
	}

	// Test variable assignments
	if !cmp.Equal(*cfg, expectedCfg) {
		t.Fatalf("Diff of expected/got config :\n %v", cmp.Diff(*cfg, expectedCfg))
	}
}
