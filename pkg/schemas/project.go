package schemas

import (
	"hash/crc32"
	"strconv"
)

// Parameters for the fetching configuration of Projects and Wildcards
type Parameters struct {
	// Fetch merge request pipelines (default: false)
	FetchMergeRequestsPipelinesRefsValue *bool `yaml:"fetch_merge_request_pipelines_refs"`

	// Maximum number for merge requests pipelines to attempt fetch on each ref discovery (default: 1)
	FetchMergeRequestsPipelinesRefsLimitValue *int `yaml:"fetch_merge_request_pipelines_refs_limit"`

	// Whether to attempt to retrieve job metrics from polled pipelines
	FetchPipelineJobMetricsValue *bool `yaml:"fetch_pipeline_job_metrics"`

	// Whether to attempt to retrieve variables included in the pipeline execution
	FetchPipelineVariablesValue *bool `yaml:"fetch_pipeline_variables"`

	// Whether to report all pipeline / job statuses, or only report the one from the last job.
	OutputSparseStatusMetricsValue *bool `yaml:"output_sparse_status_metrics"`

	// Regular expression to filter pipeline variables values to fetch (defaults to '.*')
	PipelineVariablesRegexpValue *string `yaml:"pipeline_variables_regexp"`

	// Regular expression to filter project refs to fetch (defaults to '.*')
	RefsRegexpValue *string `yaml:"refs_regexp"`
}

var (
	defaultFetchMergeRequestsPipelinesRefs      = false
	defaultFetchMergeRequestsPipelinesRefsLimit = 1
	defaultFetchPipelineJobMetrics              = false
	defaultFetchPipelineVariables               = false
	defaultOutputSparseStatusMetrics            = false
	defaultPipelineVariablesRegexp              = `.*`
	defaultRefsRegexp                           = `^(main|master)$`
)

// UpdateProjectDefaults ..
func UpdateProjectDefaults(newDefaults Parameters) {
	if newDefaults.FetchMergeRequestsPipelinesRefsValue != nil {
		defaultFetchMergeRequestsPipelinesRefs = *newDefaults.FetchMergeRequestsPipelinesRefsValue
	}

	if newDefaults.FetchMergeRequestsPipelinesRefsLimitValue != nil {
		defaultFetchMergeRequestsPipelinesRefsLimit = *newDefaults.FetchMergeRequestsPipelinesRefsLimitValue
	}

	if newDefaults.FetchPipelineJobMetricsValue != nil {
		defaultFetchPipelineJobMetrics = *newDefaults.FetchPipelineJobMetricsValue
	}

	if newDefaults.FetchPipelineVariablesValue != nil {
		defaultFetchPipelineVariables = *newDefaults.FetchPipelineVariablesValue
	}

	if newDefaults.OutputSparseStatusMetricsValue != nil {
		defaultOutputSparseStatusMetrics = *newDefaults.OutputSparseStatusMetricsValue
	}

	if newDefaults.PipelineVariablesRegexpValue != nil {
		defaultPipelineVariablesRegexp = *newDefaults.PipelineVariablesRegexpValue
	}

	if newDefaults.RefsRegexpValue != nil {
		defaultRefsRegexp = *newDefaults.RefsRegexpValue
	}
}

// Project holds information about a GitLab project
type Project struct {
	Parameters `yaml:",inline"`

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

// FetchPipelineJobMetrics ...
func (p *Project) FetchPipelineJobMetrics() bool {
	if p.FetchPipelineJobMetricsValue != nil {
		return *p.FetchPipelineJobMetricsValue
	}

	return defaultFetchPipelineJobMetrics
}

// OutputSparseStatusMetrics ...
func (p *Project) OutputSparseStatusMetrics() bool {
	if p.OutputSparseStatusMetricsValue != nil {
		return *p.OutputSparseStatusMetricsValue
	}

	return defaultOutputSparseStatusMetrics
}

// FetchPipelineVariables ...
func (p *Project) FetchPipelineVariables() bool {
	if p.FetchPipelineVariablesValue != nil {
		return *p.FetchPipelineVariablesValue
	}

	return defaultFetchPipelineVariables
}

// PipelineVariablesRegexp ...
func (p *Project) PipelineVariablesRegexp() string {
	if p.PipelineVariablesRegexpValue != nil {
		return *p.PipelineVariablesRegexpValue
	}

	return defaultPipelineVariablesRegexp
}

// RefsRegexp ...
func (p *Project) RefsRegexp() string {
	if p.RefsRegexpValue != nil {
		return *p.RefsRegexpValue
	}

	return defaultRefsRegexp
}

// FetchMergeRequestsPipelinesRefs ...
func (p *Project) FetchMergeRequestsPipelinesRefs() bool {
	if p.FetchMergeRequestsPipelinesRefsValue != nil {
		return *p.FetchMergeRequestsPipelinesRefsValue
	}

	return defaultFetchMergeRequestsPipelinesRefs
}

// FetchMergeRequestsPipelinesRefsLimit ...
func (p *Project) FetchMergeRequestsPipelinesRefsLimit() int {
	if p.FetchMergeRequestsPipelinesRefsLimitValue != nil {
		return *p.FetchMergeRequestsPipelinesRefsLimitValue
	}

	return defaultFetchMergeRequestsPipelinesRefsLimit
}
