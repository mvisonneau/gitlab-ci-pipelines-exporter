package schemas

import (
	"hash/crc32"
	"strconv"
)

var (
	defaultProjectOutputSparseStatusMetrics        = true
	defaultProjectPullRefsRegexp                   = `^(main|master)$`
	defaultProjectPullRefsFromPipelinesEnabled     = false
	defaultProjectPullRefsFromPipelinesDepth       = 100
	defaultProjectPullRefsFromMergeRequestsEnabled = false
	defaultProjectPullRefsFromMergeRequestsDepth   = 1
	defaultProjectPullPipelineJobsEnabled          = false
	defaultProjectPullPipelineVariablesEnabled     = false
	defaultProjectPullPipelineVariablesRegexp      = `.*`
)

// ProjectParameters for the fetching configuration of Projects and Wildcards
type ProjectParameters struct {
	// From handles ProjectPullParameters configuration
	Pull ProjectPull `yaml:"pull"`

	// Whether or not to export all pipeline/job statuses (being 0) or solely the one of the last job (being 1).
	OutputSparseStatusMetricsValue *bool `yaml:"output_sparse_status_metrics"`
}

// ProjectPull ..
type ProjectPull struct {
	Refs     ProjectPullRefs     `yaml:"refs"`
	Pipeline ProjectPullPipeline `yaml:"pipeline"`
}

// ProjectPullRefs ..
type ProjectPullRefs struct {
	// Regular expression to filter project refs to fetch (defaults to '.*')
	RegexpValue *string `yaml:"regexp"`

	// From handles ProjectPullRefsFromParameters configuration
	From ProjectPullRefsFrom `yaml:"from"`
}

// ProjectPullRefsFrom ..
type ProjectPullRefsFrom struct {
	// Pipelines defines whether or not to fetch refs from historical pipelines
	Pipelines ProjectPullRefsFromPipelines `yaml:"pipelines"`

	// MergeRequests defines whether or not to fetch refs from merge requests
	MergeRequests ProjectPullRefsFromMergeRequests `yaml:"merge_requests"`
}

// ProjectPullRefsFromParameters ..
type ProjectPullRefsFromParameters struct {
	EnabledValue *bool `yaml:"enabled"`
	DepthValue   *int  `yaml:"depth"`
}

// ProjectPullRefsFromPipelines ..
type ProjectPullRefsFromPipelines ProjectPullRefsFromParameters

// ProjectPullRefsFromMergeRequests ..
type ProjectPullRefsFromMergeRequests ProjectPullRefsFromParameters

// ProjectPullPipeline ..
type ProjectPullPipeline struct {
	Jobs      ProjectPullPipelineJobs      `yaml:"jobs"`
	Variables ProjectPullPipelineVariables `yaml:"variables"`
}

// ProjectPullPipelineJobs ..
type ProjectPullPipelineJobs struct {
	// Enabled set to true will pull pipeline jobs related metrics
	EnabledValue *bool `yaml:"enabled"`
}

// ProjectPullPipelineVariables ..
type ProjectPullPipelineVariables struct {
	// Enabled set to true will attempt to retrieve variables included in the pipeline
	EnabledValue *bool `yaml:"enabled"`

	// Regexp to filter pipeline variables values to fetch (defaults to '.*')
	RegexpValue *string `yaml:"regexp"`
}

// UpdateProjectDefaults ..
func UpdateProjectDefaults(d ProjectParameters) {
	if d.Pull.Refs.RegexpValue != nil {
		defaultProjectPullRefsRegexp = *d.Pull.Refs.RegexpValue
	}

	if d.Pull.Refs.From.Pipelines.EnabledValue != nil {
		defaultProjectPullRefsFromPipelinesEnabled = *d.Pull.Refs.From.Pipelines.EnabledValue
	}

	if d.Pull.Refs.From.Pipelines.DepthValue != nil {
		defaultProjectPullRefsFromPipelinesDepth = *d.Pull.Refs.From.Pipelines.DepthValue
	}

	if d.Pull.Refs.From.MergeRequests.EnabledValue != nil {
		defaultProjectPullRefsFromMergeRequestsEnabled = *d.Pull.Refs.From.MergeRequests.EnabledValue
	}

	if d.Pull.Refs.From.MergeRequests.DepthValue != nil {
		defaultProjectPullRefsFromMergeRequestsDepth = *d.Pull.Refs.From.MergeRequests.DepthValue
	}

	if d.Pull.Pipeline.Jobs.EnabledValue != nil {
		defaultProjectPullPipelineJobsEnabled = *d.Pull.Pipeline.Jobs.EnabledValue
	}

	if d.Pull.Pipeline.Variables.EnabledValue != nil {
		defaultProjectPullPipelineVariablesEnabled = *d.Pull.Pipeline.Variables.EnabledValue
	}

	if d.Pull.Pipeline.Variables.RegexpValue != nil {
		defaultProjectPullPipelineVariablesRegexp = *d.Pull.Pipeline.Variables.RegexpValue
	}
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

// OutputSparseStatusMetrics ...
func (p *ProjectParameters) OutputSparseStatusMetrics() bool {
	if p.OutputSparseStatusMetricsValue != nil {
		return *p.OutputSparseStatusMetricsValue
	}

	return defaultProjectOutputSparseStatusMetrics
}

// Regexp ...
func (p *ProjectPullRefs) Regexp() string {
	if p.RegexpValue != nil {
		return *p.RegexpValue
	}

	return defaultProjectPullRefsRegexp
}

// Enabled ...
func (p *ProjectPullRefsFromPipelines) Enabled() bool {
	if p.EnabledValue != nil {
		return *p.EnabledValue
	}

	return defaultProjectPullRefsFromPipelinesEnabled
}

// Depth ...
func (p *ProjectPullRefsFromPipelines) Depth() int {
	if p.DepthValue != nil {
		return *p.DepthValue
	}

	return defaultProjectPullRefsFromPipelinesDepth
}

// Enabled ...
func (p *ProjectPullRefsFromMergeRequests) Enabled() bool {
	if p.EnabledValue != nil {
		return *p.EnabledValue
	}

	return defaultProjectPullRefsFromMergeRequestsEnabled
}

// Depth ...
func (p *ProjectPullRefsFromMergeRequests) Depth() int {
	if p.DepthValue != nil {
		return *p.DepthValue
	}

	return defaultProjectPullRefsFromMergeRequestsDepth
}

// Enabled ...
func (p *ProjectPullPipelineJobs) Enabled() bool {
	if p.EnabledValue != nil {
		return *p.EnabledValue
	}

	return defaultProjectPullPipelineJobsEnabled
}

// Enabled ...
func (p *ProjectPullPipelineVariables) Enabled() bool {
	if p.EnabledValue != nil {
		return *p.EnabledValue
	}

	return defaultProjectPullPipelineVariablesEnabled
}

// Regexp ...
func (p *ProjectPullPipelineVariables) Regexp() string {
	if p.RegexpValue != nil {
		return *p.RegexpValue
	}

	return defaultProjectPullPipelineVariablesRegexp
}
