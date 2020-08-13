package schemas

// Parameters for the fetching configuration of Projects and Wildcards
type Parameters struct {
	// Whether to attempt to retrieve job metrics from polled pipelines
	FetchPipelineJobMetricsValue *bool `yaml:"fetch_pipeline_job_metrics"`

	// Whether to report all pipeline / job statuses, or only report the one from the last job.
	OutputSparseStatusMetricsValue *bool `yaml:"output_sparse_status_metrics"`

	// Whether to attempt to retrieve variables included in the pipeline execution
	FetchPipelineVariablesValue *bool `yaml:"fetch_pipeline_variables"`

	// Regular expression to filter pipeline variables values to fetch (defaults to '.*')
	PipelineVariablesRegexpValue *string `yaml:"pipeline_variables_regexp"`

	// Regular expression to filter project refs to fetch (defaults to '.*')
	RefsRegexpValue *string `yaml:"refs_regexp"`

	// Fetch merge request pipelines (default: false)
	FetchMergeRequestsPipelinesRefsValue *bool `yaml:"fetch_merge_request_pipelines_refs"`

	// Maximum number for merge requests pipelines to attempt fetch on each ref discovery (default: 1)
	FetchMergeRequestsPipelinesRefsLimitValue *int `yaml:"fetch_merge_request_pipelines_refs_limit"`
}

const (
	defaultFetchPipelineJobMetrics              = false
	defaultOutputSparseStatusMetrics            = false
	defaultFetchPipelineVariables               = false
	defaultRefsRegexp                           = `^(main|master)$`
	defaultPipelineVariablesRegexp              = `.*`
	defaultFetchMergeRequestsPipelinesRefs      = false
	defaultFetchMergeRequestsPipelinesRefsLimit = 1
)

// Project holds information about a GitLab project
type Project struct {
	Parameters `yaml:",inline"`

	Name string `yaml:"name"`
}

// FetchPipelineJobMetrics ...
func (p *Project) FetchPipelineJobMetrics(cfg *Config) bool {
	if p.FetchPipelineJobMetricsValue != nil {
		return *p.FetchPipelineJobMetricsValue
	}

	if cfg.Defaults.FetchPipelineJobMetricsValue != nil {
		return *cfg.Defaults.FetchPipelineJobMetricsValue
	}

	return defaultFetchPipelineJobMetrics
}

// OutputSparseStatusMetrics ...
func (p *Project) OutputSparseStatusMetrics(cfg *Config) bool {
	if p.OutputSparseStatusMetricsValue != nil {
		return *p.OutputSparseStatusMetricsValue
	}

	if cfg.Defaults.OutputSparseStatusMetricsValue != nil {
		return *cfg.Defaults.OutputSparseStatusMetricsValue
	}

	return defaultOutputSparseStatusMetrics
}

// FetchPipelineVariables ...
func (p *Project) FetchPipelineVariables(cfg *Config) bool {
	if p.FetchPipelineVariablesValue != nil {
		return *p.FetchPipelineVariablesValue
	}

	if cfg.Defaults.FetchPipelineVariablesValue != nil {
		return *cfg.Defaults.FetchPipelineVariablesValue
	}

	return defaultFetchPipelineVariables
}

// PipelineVariablesRegexp ...
func (p *Project) PipelineVariablesRegexp(cfg *Config) string {
	if p.PipelineVariablesRegexpValue != nil {
		return *p.PipelineVariablesRegexpValue
	}

	if cfg.Defaults.PipelineVariablesRegexpValue != nil {
		return *cfg.Defaults.PipelineVariablesRegexpValue
	}

	return defaultPipelineVariablesRegexp
}

// RefsRegexp ...
func (p *Project) RefsRegexp(cfg *Config) string {
	if p.RefsRegexpValue != nil {
		return *p.RefsRegexpValue
	}

	if cfg.Defaults.RefsRegexpValue != nil {
		return *cfg.Defaults.RefsRegexpValue
	}

	return defaultRefsRegexp
}

// FetchMergeRequestsPipelinesRefs ...
func (p *Project) FetchMergeRequestsPipelinesRefs(cfg *Config) bool {
	if p.FetchMergeRequestsPipelinesRefsValue != nil {
		return *p.FetchMergeRequestsPipelinesRefsValue
	}

	if cfg.Defaults.FetchMergeRequestsPipelinesRefsValue != nil {
		return *cfg.Defaults.FetchMergeRequestsPipelinesRefsValue
	}

	return defaultFetchMergeRequestsPipelinesRefs
}

// FetchMergeRequestsPipelinesRefsLimit ...
func (p *Project) FetchMergeRequestsPipelinesRefsLimit(cfg *Config) int {
	if p.FetchMergeRequestsPipelinesRefsLimitValue != nil {
		return *p.FetchMergeRequestsPipelinesRefsLimitValue
	}

	if cfg.Defaults.FetchMergeRequestsPipelinesRefsLimitValue != nil {
		return *cfg.Defaults.FetchMergeRequestsPipelinesRefsLimitValue
	}

	return defaultFetchMergeRequestsPipelinesRefsLimit
}
