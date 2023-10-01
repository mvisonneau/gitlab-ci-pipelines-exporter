package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefKey(t *testing.T) {
	assert.Equal(t, RefKey("1690074537"), NewRef(
		NewProject("foo/bar"),
		RefKindBranch,
		"baz",
	).Key())
}

func TestRefsCount(t *testing.T) {
	assert.Equal(t, 2, Refs{
		RefKey("foo"): Ref{},
		RefKey("bar"): Ref{},
	}.Count())
}

func TestRefDefaultLabelsValues(t *testing.T) {
	p := NewProject("foo/bar")
	p.Topics = "amazing,project"
	ref := Ref{
		Project: p,
		Kind:    RefKindBranch,
		Name:    "feature",
		LatestPipeline: Pipeline{
			Variables: "blah",
			Source:    "schedule",
		},
		LatestJobs: make(Jobs),
	}

	expectedValue := map[string]string{
		"kind":      "branch",
		"project":   "foo/bar",
		"ref":       "feature",
		"topics":    "amazing,project",
		"variables": "blah",
		"source":    "schedule",
	}

	assert.Equal(t, expectedValue, ref.DefaultLabelsValues())
}

func TestNewRef(t *testing.T) {
	p := NewProject("foo/bar")
	p.Topics = "bar,baz"
	p.OutputSparseStatusMetrics = false
	p.Pull.Pipeline.Jobs.Enabled = true
	p.Pull.Pipeline.Jobs.FromChildPipelines.Enabled = false
	p.Pull.Pipeline.Jobs.RunnerDescription.Enabled = false
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.Variables.Regexp = `.*`
	p.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp = `.*`

	expectedValue := Ref{
		Project:    p,
		Kind:       RefKindTag,
		Name:       "v0.0.7",
		LatestJobs: make(Jobs),
	}

	assert.Equal(t, expectedValue, NewRef(
		p,
		RefKindTag,
		"v0.0.7",
	))
}

func TestGetMergeRequestIIDFromRefName(t *testing.T) {
	name, err := GetMergeRequestIIDFromRefName("1234")
	assert.NoError(t, err)
	assert.Equal(t, "1234", name)

	name, err = GetMergeRequestIIDFromRefName("refs/merge-requests/5678/head")
	assert.NoError(t, err)
	assert.Equal(t, "5678", name)

	name, err = GetMergeRequestIIDFromRefName("20.0.1")
	assert.Error(t, err)
	assert.Equal(t, "20.0.1", name)

	name, err = GetMergeRequestIIDFromRefName("x")
	assert.Error(t, err)
	assert.Equal(t, "x", name)
}
