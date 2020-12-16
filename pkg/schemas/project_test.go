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

	assert.Equal(t, ProjectKey("2356372769"), p.Key())
}

func NewTestProjectVariables() (cfg *Config, project *Project) {
	cfg = &Config{}

	project = &Project{
		Name: "foo",
	}

	return
}

func TestOutputSparseStatusMetrics(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())

	cfg.ProjectDefaults.OutputSparseStatusMetricsValue = pointy.Bool(!defaultProjectOutputSparseStatusMetrics)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())

	project.OutputSparseStatusMetricsValue = pointy.Bool(defaultProjectOutputSparseStatusMetrics)
	assert.Equal(t, defaultProjectOutputSparseStatusMetrics, project.OutputSparseStatusMetrics())
}

func TestPullEnvironmentsFromProjectsEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullEnvironmentsEnabled, project.Pull.Environments.Enabled())

	cfg.ProjectDefaults.Pull.Environments.EnabledValue = pointy.Bool(!defaultProjectPullEnvironmentsEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullEnvironmentsEnabled, project.Pull.Environments.Enabled())

	project.Pull.Environments.EnabledValue = pointy.Bool(defaultProjectPullEnvironmentsEnabled)
	assert.Equal(t, defaultProjectPullEnvironmentsEnabled, project.Pull.Environments.Enabled())
}

func TestPullEnvironmentsFromProjectsNameRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullEnvironmentsNameRegexp, project.Pull.Environments.NameRegexp())

	cfg.ProjectDefaults.Pull.Environments.NameRegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, "foo", project.Pull.Environments.NameRegexp())

	project.Pull.Environments.NameRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.Pull.Environments.NameRegexp())
}

func TestPullEnvironmentsFromProjectsTagsRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullEnvironmentsTagsRegexp, project.Pull.Environments.TagsRegexp())

	cfg.ProjectDefaults.Pull.Environments.TagsRegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, "foo", project.Pull.Environments.TagsRegexp())

	project.Pull.Environments.TagsRegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.Pull.Environments.TagsRegexp())
}

func TestPullRefsRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsRegexp, project.Pull.Refs.Regexp())

	cfg.ProjectDefaults.Pull.Refs.RegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, "foo", project.Pull.Refs.Regexp())

	project.Pull.Refs.RegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.Pull.Refs.Regexp())
}

func TestPullRefsMaxAgeSeconds(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsMaxAgeSeconds, project.Pull.Refs.MaxAgeSeconds())

	cfg.ProjectDefaults.Pull.Refs.MaxAgeSecondsValue = pointy.Uint(1)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, uint(1), project.Pull.Refs.MaxAgeSeconds())

	project.Pull.Refs.MaxAgeSecondsValue = pointy.Uint(2)
	assert.Equal(t, uint(2), project.Pull.Refs.MaxAgeSeconds())
}

func TestPullRefsFromPipelinesEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsFromPipelinesEnabled, project.Pull.Refs.From.Pipelines.Enabled())

	cfg.ProjectDefaults.Pull.Refs.From.Pipelines.EnabledValue = pointy.Bool(!defaultProjectPullRefsFromPipelinesEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullRefsFromPipelinesEnabled, project.Pull.Refs.From.Pipelines.Enabled())

	project.Pull.Refs.From.Pipelines.EnabledValue = pointy.Bool(defaultProjectPullRefsFromPipelinesEnabled)
	assert.Equal(t, defaultProjectPullRefsFromPipelinesEnabled, project.Pull.Refs.From.Pipelines.Enabled())
}

func TestPullRefsFromPipelinesDepth(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsFromPipelinesDepth, project.Pull.Refs.From.Pipelines.Depth())

	cfg.ProjectDefaults.Pull.Refs.From.Pipelines.DepthValue = pointy.Int(1)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, 1, project.Pull.Refs.From.Pipelines.Depth())

	project.Pull.Refs.From.Pipelines.DepthValue = pointy.Int(2)
	assert.Equal(t, 2, project.Pull.Refs.From.Pipelines.Depth())
}

func TestPullRefsFromMergeRequestsEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsEnabled, project.Pull.Refs.From.MergeRequests.Enabled())

	cfg.ProjectDefaults.Pull.Refs.From.MergeRequests.EnabledValue = pointy.Bool(!defaultProjectPullRefsFromMergeRequestsEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsEnabled, project.Pull.Refs.From.MergeRequests.Enabled())

	project.Pull.Refs.From.MergeRequests.EnabledValue = pointy.Bool(defaultProjectPullRefsFromPipelinesEnabled)
	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsEnabled, project.Pull.Refs.From.MergeRequests.Enabled())
}

func TestPullRefsFromMergeRequestsDepth(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullRefsFromMergeRequestsDepth, project.Pull.Refs.From.MergeRequests.Depth())

	cfg.ProjectDefaults.Pull.Refs.From.MergeRequests.DepthValue = pointy.Int(1)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, 1, project.Pull.Refs.From.MergeRequests.Depth())

	project.Pull.Refs.From.MergeRequests.DepthValue = pointy.Int(2)
	assert.Equal(t, 2, project.Pull.Refs.From.MergeRequests.Depth())
}

func TestPullPipelineJobsEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullPipelineJobsEnabled, project.Pull.Pipeline.Jobs.Enabled())

	cfg.ProjectDefaults.Pull.Pipeline.Jobs.EnabledValue = pointy.Bool(!defaultProjectPullPipelineJobsEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullPipelineJobsEnabled, project.Pull.Pipeline.Jobs.Enabled())

	project.Pull.Pipeline.Jobs.EnabledValue = pointy.Bool(defaultProjectPullPipelineJobsEnabled)
	assert.Equal(t, defaultProjectPullPipelineJobsEnabled, project.Pull.Pipeline.Jobs.Enabled())
}

func TestPullPipelineJobsFromChildPipelinesEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullPipelineJobsFromChildPipelinesEnabled, project.Pull.Pipeline.Jobs.FromChildPipelines.Enabled())

	cfg.ProjectDefaults.Pull.Pipeline.Jobs.FromChildPipelines.EnabledValue = pointy.Bool(!defaultProjectPullPipelineJobsFromChildPipelinesEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullPipelineJobsFromChildPipelinesEnabled, project.Pull.Pipeline.Jobs.FromChildPipelines.Enabled())

	project.Pull.Pipeline.Jobs.FromChildPipelines.EnabledValue = pointy.Bool(defaultProjectPullPipelineJobsFromChildPipelinesEnabled)
	assert.Equal(t, defaultProjectPullPipelineJobsFromChildPipelinesEnabled, project.Pull.Pipeline.Jobs.FromChildPipelines.Enabled())
}

func TestPullPipelineVariablesEnabled(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullPipelineVariablesEnabled, project.Pull.Pipeline.Variables.Enabled())

	cfg.ProjectDefaults.Pull.Pipeline.Variables.EnabledValue = pointy.Bool(!defaultProjectPullPipelineVariablesEnabled)
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, defaultProjectPullPipelineVariablesEnabled, project.Pull.Pipeline.Variables.Enabled())

	project.Pull.Pipeline.Variables.EnabledValue = pointy.Bool(defaultProjectPullPipelineVariablesEnabled)
	assert.Equal(t, defaultProjectPullPipelineVariablesEnabled, project.Pull.Pipeline.Variables.Enabled())
}

func TestPullPipelineVariablesRegexp(t *testing.T) {
	cfg, project := NewTestProjectVariables()
	assert.Equal(t, defaultProjectPullPipelineVariablesRegexp, project.Pull.Pipeline.Variables.Regexp())

	cfg.ProjectDefaults.Pull.Pipeline.Variables.RegexpValue = pointy.String("foo")
	UpdateProjectDefaults(cfg.ProjectDefaults)
	assert.Equal(t, "foo", project.Pull.Pipeline.Variables.Regexp())

	project.Pull.Pipeline.Variables.RegexpValue = pointy.String("bar")
	assert.Equal(t, "bar", project.Pull.Pipeline.Variables.Regexp())
}
