package schemas

import (
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func TestProjectKey(t *testing.T) {
	p := Project{
		Name: "foo",
	}

	assert.Equal(t, ProjectKey("C-7Hteo_D9vJXQ3UfzxbwnXaijM="), p.Key())
}

func NewTestProjectVariables() (cfg *Config, project *Project) {
	cfg = &Config{}

	project = &Project{
		Name: "foo",
	}

	return
}

func TestFetchPipelineJobMetrics(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics())

	cfg.Defaults.FetchPipelineJobMetricsValue = pointy.Bool(!defaultFetchPipelineJobMetrics)
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics())

	project.FetchPipelineJobMetricsValue = pointy.Bool(defaultFetchPipelineJobMetrics)
	assert.Equal(t, defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics())
}

func TestOutputSparseStatusMetrics(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())

	cfg.Defaults.OutputSparseStatusMetricsValue = pointy.Bool(!defaultOutputSparseStatusMetrics)
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())

	project.OutputSparseStatusMetricsValue = pointy.Bool(defaultOutputSparseStatusMetrics)
	assert.Equal(t, defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())
}

func TestFetchPipelineVariables(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchPipelineVariables, project.FetchPipelineVariables())

	cfg.Defaults.FetchPipelineVariablesValue = pointy.Bool(!defaultFetchPipelineVariables)
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, defaultFetchPipelineVariables, project.FetchPipelineVariables())

	project.FetchPipelineVariablesValue = pointy.Bool(defaultFetchPipelineVariables)
	assert.Equal(t, defaultFetchPipelineVariables, project.FetchPipelineVariables())
}

func TestPipelineVariablesRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultPipelineVariablesRegexp, project.PipelineVariablesRegexp())

	cfg.Defaults.PipelineVariablesRegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, "foo", project.PipelineVariablesRegexp())

	project.PipelineVariablesRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.PipelineVariablesRegexp())
}

func TestRefsRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultRefsRegexp, project.RefsRegexp())

	cfg.Defaults.RefsRegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, "foo", project.RefsRegexp())

	project.RefsRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.RefsRegexp())
}

func TestFetchMergeRequestsPipelinesRefs(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchMergeRequestsPipelinesRefs, project.FetchMergeRequestsPipelinesRefs())

	cfg.Defaults.FetchMergeRequestsPipelinesRefsValue = pointy.Bool(!defaultFetchMergeRequestsPipelinesRefs)
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, defaultFetchMergeRequestsPipelinesRefs, project.FetchMergeRequestsPipelinesRefs())

	project.FetchMergeRequestsPipelinesRefsValue = pointy.Bool(defaultFetchMergeRequestsPipelinesRefs)
	assert.Equal(t, defaultFetchMergeRequestsPipelinesRefs, project.FetchMergeRequestsPipelinesRefs())
}

func TestFetchMergeRequestsPipelinesRefsLimit(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchMergeRequestsPipelinesRefsLimit, project.FetchMergeRequestsPipelinesRefsLimit())

	cfg.Defaults.FetchMergeRequestsPipelinesRefsLimitValue = pointy.Int(10)
	UpdateProjectDefaults(cfg.Defaults)
	assert.Equal(t, 10, project.FetchMergeRequestsPipelinesRefsLimit())

	project.FetchMergeRequestsPipelinesRefsLimitValue = pointy.Int(20)
	assert.Equal(t, 20, project.FetchMergeRequestsPipelinesRefsLimit())
}
