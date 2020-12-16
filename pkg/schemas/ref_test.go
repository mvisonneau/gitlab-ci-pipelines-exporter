package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefKey(t *testing.T) {
	ref := Ref{
		Kind:        RefKindBranch,
		ProjectName: "foo/bar",
		Name:        "baz",
	}

	assert.Equal(t, RefKey("1690074537"), ref.Key())
}

func TestRefsCount(t *testing.T) {
	assert.Equal(t, 2, Refs{
		RefKey("foo"): Ref{},
		RefKey("bar"): Ref{},
	}.Count())
}

func TestRefDefaultLabelsValues(t *testing.T) {
	ref := Ref{
		Kind:        RefKindBranch,
		ProjectName: "foo/bar",
		Name:        "feature",
		Topics:      "amazing,project",
		LatestPipeline: Pipeline{
			Variables: "blah",
		},
		LatestJobs: make(Jobs),
	}

	expectedValue := map[string]string{
		"kind":      "branch",
		"project":   "foo/bar",
		"ref":       "feature",
		"topics":    "amazing,project",
		"variables": "blah",
	}

	assert.Equal(t, expectedValue, ref.DefaultLabelsValues())
}

func TestNewRef(t *testing.T) {
	expectedValue := Ref{
		Kind:        RefKindTag,
		ProjectName: "foo/bar",
		Name:        "v0.0.7",
		Topics:      "bar,baz",
		LatestJobs:  make(Jobs),

		OutputSparseStatusMetrics:                 true,
		PullPipelineJobsEnabled:                   true,
		PullPipelineJobsFromChildPipelinesEnabled: false,
		PullPipelineVariablesEnabled:              true,
		PullPipelineVariablesRegexp:               ".*",
	}

	assert.Equal(t, expectedValue, NewRef(
		RefKindTag,
		"foo/bar",
		"v0.0.7",
		"bar,baz",
		true,
		true,
		false,
		true,
		".*",
	))
}
