package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProject(t *testing.T) {
	p := Project{}

	p.Name = "foo/bar"

	p.OutputSparseStatusMetrics = true

	p.Pull.Environments.Regexp = `.*`
	p.Pull.Environments.ExcludeStopped = true

	p.Pull.Refs.Branches.Enabled = true
	p.Pull.Refs.Branches.Regexp = `^(?:main|master)$`
	p.Pull.Refs.Branches.ExcludeDeleted = true

	p.Pull.Refs.Tags.Enabled = true
	p.Pull.Refs.Tags.Regexp = `.*`
	p.Pull.Refs.Tags.ExcludeDeleted = true

	p.Pull.Pipeline.Jobs.FromChildPipelines.Enabled = true
	p.Pull.Pipeline.Jobs.RunnerDescription.Enabled = true
	p.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp = `shared-runners-manager-(\d*)\.gitlab\.com`
	p.Pull.Pipeline.Variables.Regexp = `.*`

	assert.Equal(t, p, NewProject("foo/bar"))
}
