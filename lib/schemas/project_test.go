package schemas

import (
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func NewTestProjectVariables() (cfg *Config, project *Project) {
	cfg = &Config{}

	project = &Project{
		Name: "foo",
	}

	return
}

func TestFetchPipelineJobMetrics(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics(cfg))

	cfg.Defaults.FetchPipelineJobMetricsValue = pointy.Bool(!defaultFetchPipelineJobMetrics)
	assert.Equal(t, !defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics(cfg))

	project.FetchPipelineJobMetricsValue = pointy.Bool(defaultFetchPipelineJobMetrics)
	assert.Equal(t, defaultFetchPipelineJobMetrics, project.FetchPipelineJobMetrics(cfg))
}

func TestOutputSparseStatusMetrics(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics(cfg))

	cfg.Defaults.OutputSparseStatusMetricsValue = pointy.Bool(!defaultOutputSparseStatusMetrics)
	assert.Equal(t, !defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics(cfg))

	project.OutputSparseStatusMetricsValue = pointy.Bool(defaultOutputSparseStatusMetrics)
	assert.Equal(t, defaultOutputSparseStatusMetrics, project.OutputSparseStatusMetrics(cfg))
}

func TestFetchPipelineVariables(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultFetchPipelineVariables, project.FetchPipelineVariables(cfg))

	cfg.Defaults.FetchPipelineVariablesValue = pointy.Bool(!defaultFetchPipelineVariables)
	assert.Equal(t, !defaultFetchPipelineVariables, project.FetchPipelineVariables(cfg))

	project.FetchPipelineVariablesValue = pointy.Bool(defaultFetchPipelineVariables)
	assert.Equal(t, defaultFetchPipelineVariables, project.FetchPipelineVariables(cfg))
}

func TestPipelineVariablesRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultPipelineVariablesRegexp, project.PipelineVariablesRegexp(cfg))

	cfg.Defaults.PipelineVariablesRegexpValue = pointy.String("foo")
	assert.Equal(t, "foo", project.PipelineVariablesRegexp(cfg))

	project.PipelineVariablesRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.PipelineVariablesRegexp(cfg))
}

func TestRefsRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultRefsRegexp, project.RefsRegexp(cfg))

	cfg.Defaults.RefsRegexpValue = pointy.String("foo")
	assert.Equal(t, "foo", project.RefsRegexp(cfg))

	project.RefsRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.RefsRegexp(cfg))
}
