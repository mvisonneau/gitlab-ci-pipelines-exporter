package config

import (
	"hash/crc32"
	"strconv"

	"github.com/creasty/defaults"
)

// ProjectParameters for the fetching configuration of Projects and Wildcards
type ProjectParameters struct {
	// From handles ProjectPullParameters configuration
	Pull ProjectPull `yaml:"pull"`

	// Whether or not to export all pipeline/job statuses (being 0) or solely the one of the last job (being 1).
	OutputSparseStatusMetrics bool `default:"true" yaml:"output_sparse_status_metrics"`
}

// ProjectPull ..
type ProjectPull struct {
	Environments ProjectPullEnvironments `yaml:"environments"`
	Refs         ProjectPullRefs         `yaml:"refs"`
	Pipeline     ProjectPullPipeline     `yaml:"pipeline"`
}

// ProjectPullEnvironments ..
type ProjectPullEnvironments struct {
	// Whether to pull environments/deployments or not for this project
	Enabled bool `default:"false" yaml:"enabled"`

	// Regular expression to filter environments to fetch by their names
	Regexp string `default:".*" yaml:"regexp"`
}

// ProjectPullRefs ..
type ProjectPullRefs struct {
	// Regular expression to filter refs to fetch
	Regexp string `default:"^(main|master)$" yaml:"regexp"`

	// If the age of the most recent pipeline for the ref is greater than this value, the ref won't get exported
	MaxAgeSeconds uint `default:"0" yaml:"max_age_seconds"`

	// From handles ProjectPullRefsFromParameters configuration
	From ProjectPullRefsFrom `yaml:"from"`
}

// ProjectPullRefsFrom ..
type ProjectPullRefsFrom struct {
	// Pipelines defines whether or not to fetch refs from historical pipelines
	Pipelines struct {
		Enabled bool `default:"false" yaml:"enabled"`
		Depth   uint `default:"50" yaml:"depth"`
	} `yaml:"pipelines"`

	// MergeRequests defines whether or not to fetch refs from merge requests
	MergeRequests struct {
		Enabled bool `default:"false" yaml:"enabled"`
		Depth   uint `default:"10" yaml:"depth"`
	} `yaml:"merge_requests"`
}

// ProjectPullPipeline ..
type ProjectPullPipeline struct {
	Jobs      ProjectPullPipelineJobs      `yaml:"jobs"`
	Variables ProjectPullPipelineVariables `yaml:"variables"`
}

// ProjectPullPipelineJobs ..
type ProjectPullPipelineJobs struct {
	// Enabled set to true will pull pipeline jobs related metrics
	Enabled bool `default:"false" yaml:"enabled"`

	// Pull pipeline jobs from child/downstream pipelines
	FromChildPipelines ProjectPullPipelineJobsFromChildPipelines `yaml:"from_child_pipelines"`

	// Configure the export of the runner description which ran the job
	RunnerDescription ProjectPullPipelineJobsRunnerDescription `yaml:"runner_description"`
}

// ProjectPullPipelineJobsFromChildPipelines ..
type ProjectPullPipelineJobsFromChildPipelines struct {
	// Enabled set to true will pull pipeline jobs from child/downstream pipelines related metrics
	Enabled bool `default:"true" yaml:"enabled"`
}

// ProjectPullPipelineJobsRunnerDescription ..
type ProjectPullPipelineJobsRunnerDescription struct {
	// Enabled set to true will export the description of the runner which ran the job
	Enabled bool `default:"true" yaml:"enabled"`

	// Regular expression to be able to reduce the cardinality of the exported value when necessary
	AggregationRegexp string `default:"shared-runners-manager-(\\d*)\\.gitlab\\.com" yaml:"aggregation_regexp"`
}

// ProjectPullPipelineVariables ..
type ProjectPullPipelineVariables struct {
	// Enabled set to true will attempt to retrieve variables included in the pipeline
	Enabled bool `default:"false" yaml:"enabled"`

	// Regexp to filter pipeline variables values to fetch
	Regexp string `default:".*" yaml:"regexp"`
}

// Project holds information about a GitLab project
type Project struct {
	// ProjectParameters holds parameters specific to this project
	ProjectParameters `yaml:",inline"`

	// Name is actually what is commonly referred as path_with_namespace on GitLab
	Name string `yaml:"name"`
}

// ProjectKey ..
type ProjectKey string

// Key ..
func (p Project) Key() ProjectKey {
	return ProjectKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(p.Name)))))
}

// Projects ..
type Projects map[ProjectKey]Project

// NewProject returns a new project composed with the default parameters
func NewProject(name string) (p Project) {
	defaults.MustSet(&p)
	p.Name = name
	return
}
