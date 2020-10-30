package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestRefKey(t *testing.T) {
	ref := Ref{
		PathWithNamespace: "foo/bar",
		Ref:               "baz",
	}

	assert.Equal(t, RefKey("732621927"), ref.Key())
}

func TestRefsCount(t *testing.T) {
	assert.Equal(t, 2, Refs{
		RefKey("foo"): Ref{},
		RefKey("bar"): Ref{},
	}.Count())
}

func TestRefDefaultLabelsValues(t *testing.T) {
	ref := Ref{
		PathWithNamespace:           "foo/bar",
		Topics:                      "amazing",
		Ref:                         "feature",
		Kind:                        RefKindBranch,
		MostRecentPipelineVariables: "blah",
	}

	expectedValue := map[string]string{
		"project":   "foo/bar",
		"topics":    "amazing",
		"ref":       "feature",
		"kind":      "branch",
		"variables": "blah",
	}

	assert.Equal(t, expectedValue, ref.DefaultLabelsValues())
}

func TestNewRef(t *testing.T) {
	p := Project{
		Name: "foo/bar",
	}

	gp := &goGitlab.Project{
		ID:                1,
		PathWithNamespace: "foo/bar",
		TagList:           []string{"baz", "yolo"},
	}

	expectedValue := Ref{
		Project: Project{
			Name: "foo/bar",
		},
		PathWithNamespace: "foo/bar",
		Kind:              RefKindTag,
		ID:                1,
		Topics:            "baz,yolo",
		Ref:               "v0.0.7",
		Jobs:              make(map[string]goGitlab.Job),
	}

	assert.Equal(t, expectedValue, NewRef(p, gp, "v0.0.7", RefKindTag))
}
