package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWildcard(t *testing.T) {
	w := Wildcard{}

	w.OutputSparseStatusMetrics = true

	w.Pull.Environments.Regexp = `.*`
	w.Pull.Environments.ExcludeStopped = true

	w.Pull.Refs.Branches.Enabled = true
	w.Pull.Refs.Branches.Regexp = `^(?:main|master)$`
	w.Pull.Refs.Branches.ExcludeDeleted = true

	w.Pull.Refs.Tags.Enabled = true
	w.Pull.Refs.Tags.Regexp = `.*`
	w.Pull.Refs.Tags.ExcludeDeleted = true

	w.Pull.Pipeline.Jobs.FromChildPipelines.Enabled = true
	w.Pull.Pipeline.Jobs.RunnerDescription.Enabled = true
	w.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp = `shared-runners-manager-(\d*)\.gitlab\.com`
	w.Pull.Pipeline.Variables.Regexp = `.*`

	assert.Equal(t, w, NewWildcard())
}
