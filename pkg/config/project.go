package config

import (
	"github.com/creasty/defaults"
)

// ProjectParameters for the fetching configuration of Projects and Wildcards.
type ProjectParameters struct {
	// From handles ProjectPullParameters configuration.
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

	// Prevent exporting metrics for stopped environments
	ExcludeStopped bool `default:"true" yaml:"exclude_stopped"`
}

// ProjectPullRefs ..
type ProjectPullRefs struct {
	// Configuration for pulling branches
	Branches ProjectPullRefsBranches `yaml:"branches"`

	// Configuration for pulling tags
	Tags ProjectPullRefsTags `yaml:"tags"`

	// Configuration for pulling merge requests
	MergeRequests ProjectPullRefsMergeRequests `yaml:"merge_requests"`
}

// ProjectPullRefsBranches ..
type ProjectPullRefsBranches struct {
	// Monitor pipelines related to project branches
	Enabled bool `default:"true" yaml:"enabled"`

	// Filter for branches to include
	Regexp string `default:"^(?:main|master)$" yaml:"regexp"`

	// Only keep most 'n' recently updated branches
	MostRecent uint `default:"0" yaml:"most_recent"`

	// If the most recent pipeline for the branch was last updated at
	// time greater than this value the metrics won't be exported
	MaxAgeSeconds uint `default:"0" yaml:"max_age_seconds"`

	// Prevent exporting metrics for deleted branches
	ExcludeDeleted bool `default:"true" yaml:"exclude_deleted"`
}

// ProjectPullRefsTags ..
type ProjectPullRefsTags struct {
	// Monitor pipelines related to project tags.
	Enabled bool `default:"true" yaml:"enabled"`

	// Filter for tags to include.
	Regexp string `default:".*" yaml:"regexp"`

	// Only keep most 'n' recently updated tags.
	MostRecent uint `default:"0" yaml:"most_recent"`

	// If the most recent pipeline for the tag was last updated at
	// time greater than this value the metrics won't be exported.
	MaxAgeSeconds uint `default:"0" yaml:"max_age_seconds"`

	// Prevent exporting metrics for deleted tags.
	ExcludeDeleted bool `default:"true" yaml:"exclude_deleted"`
}

// ProjectPullRefsMergeRequests ..
type ProjectPullRefsMergeRequests struct {
	// Monitor pipelines related to project merge requests.
	Enabled bool `yaml:"enabled"`

	// Only keep most 'n' recently updated merge requests.
	MostRecent uint `default:"0" yaml:"most_recent"`

	// If the most recent pipeline for the merge request was last updated at
	// time greater than this value the metrics won't be exported.
	MaxAgeSeconds uint `default:"0" yaml:"max_age_seconds"`
}

// ProjectPullPipeline ..
type ProjectPullPipeline struct {
	Jobs        ProjectPullPipelineJobs        `yaml:"jobs"`
	Variables   ProjectPullPipelineVariables   `yaml:"variables"`
	TestReports ProjectPullPipelineTestReports `yaml:"test_reports"`
}

// ProjectPullPipelineJobs ..
type ProjectPullPipelineJobs struct {
	// Enabled set to true will pull pipeline jobs related metrics.
	Enabled bool `default:"false" yaml:"enabled"`

	// Pull pipeline jobs from child/downstream pipelines.
	FromChildPipelines ProjectPullPipelineJobsFromChildPipelines `yaml:"from_child_pipelines"`

	// Configure the export of the runner description which ran the job.
	RunnerDescription ProjectPullPipelineJobsRunnerDescription `yaml:"runner_description"`
}

// ProjectPullPipelineJobsFromChildPipelines ..
type ProjectPullPipelineJobsFromChildPipelines struct {
	// Enabled set to true will pull pipeline jobs from child/downstream pipelines related metrics.
	Enabled bool `default:"true" yaml:"enabled"`
}

// ProjectPullPipelineJobsRunnerDescription ..
type ProjectPullPipelineJobsRunnerDescription struct {
	// Enabled set to true will export the description of the runner which ran the job.
	Enabled bool `default:"true" yaml:"enabled"`

	// Regular expression to be able to reduce the cardinality of the exported value when necessary.
	AggregationRegexp string `default:"shared-runners-manager-(\\d*)\\.gitlab\\.com" yaml:"aggregation_regexp"`
}

// ProjectPullPipelineVariables ..
type ProjectPullPipelineVariables struct {
	// Enabled set to true will attempt to retrieve variables included in the pipeline.
	Enabled bool `default:"false" yaml:"enabled"`

	// Regexp to filter pipeline variables values to fetch.
	Regexp string `default:".*" yaml:"regexp"`
}

// ProjectPullPipelineTestReports ..
type ProjectPullPipelineTestReports struct {
	// Enabled set to true will attempt to retrieve the test report included in the pipeline.
	Enabled            bool                                             `default:"false" yaml:"enabled"`
	FromChildPipelines ProjectPullPipelineTestReportsFromChildPipelines `yaml:"from_child_pipelines"`
	TestCases          ProjectPullPipelineTestReportsTestCases          `yaml:"test_cases"`
}

// ProjectPullPipelineJobsFromChildPipelines ..
type ProjectPullPipelineTestReportsFromChildPipelines struct {
	// Enabled set to true will pull pipeline jobs from child/downstream pipelines related metrics.
	Enabled bool `default:"false" yaml:"enabled"`
}

// ProjectPullPipelineTestCases ..
type ProjectPullPipelineTestReportsTestCases struct {
	// Enabled set to true will attempt to retrieve the test report included in the pipeline.
	Enabled bool `default:"false" yaml:"enabled"`
}

// Project holds information about a GitLab project.
type Project struct {
	// ProjectParameters holds parameters specific to this project.
	ProjectParameters `yaml:",inline"`

	// Name is actually what is commonly referred as path_with_namespace on GitLab.
	Name string `yaml:"name"`
}

// Projects ..
type Projects []Project

// NewProject returns a new project composed with the default parameters.
func NewProject(name string) (p Project) {
	defaults.MustSet(&p)
	p.Name = name

	return
}
