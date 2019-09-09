package cmd

type config struct {
	Gitlab struct {
		URL           string
		Token         string
		SkipTLSVerify bool `yaml:"skip_tls_verify"`
	}

	ProjectsPollingIntervalSeconds     int `yaml:"projects_polling_interval_seconds"`
	RefsPollingIntervalSeconds         int `yaml:"refs_polling_interval_seconds"`
	PipelinesPollingIntervalSeconds    int `yaml:"pipelines_polling_interval_seconds"`
	PipelinesMaxPollingIntervalSeconds int `yaml:"pipelines_max_polling_interval_seconds"`

	DefaultRefsRegexp string `yaml:"default_refs"`

	Projects  []project
	Wildcards []wildcard
}

type project struct {
	Name string
	Refs string
}

type wildcard struct {
	Search string
	Owner  struct {
		Name string
		Kind string
	}
	Refs string
}

var cfg config
