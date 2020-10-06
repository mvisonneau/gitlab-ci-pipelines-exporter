package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestProjectRefKey(t *testing.T) {
	pr := ProjectRef{
		Project: Project{
			Name: "foo",
		},
		Ref: "bar",
	}

	assert.Equal(t, ProjectRefKey("iEPX-SQWIR3p67lj_0zigSWTKHg="), pr.Key())
}

func TestProjectRefsCount(t *testing.T) {
	assert.Equal(t, 2, ProjectsRefs{
		ProjectRefKey("foo"): ProjectRef{},
		ProjectRefKey("bar"): ProjectRef{},
	}.Count())
}

func TestProjectRefDefaultLabelsValues(t *testing.T) {
	pr := ProjectRef{
		PathWithNamespace:           "foo/bar",
		Topics:                      "amazing",
		Ref:                         "feature",
		Kind:                        ProjectRefKindBranch,
		MostRecentPipelineVariables: "blah",
	}

	expectedValue := map[string]string{
		"project":   "foo/bar",
		"topics":    "amazing",
		"ref":       "feature",
		"kind":      "branch",
		"variables": "blah",
	}

	assert.Equal(t, expectedValue, pr.DefaultLabelsValues())
}

func TestNewProjectRef(t *testing.T) {
	p := Project{
		Name: "foo/bar",
	}

	gp := &goGitlab.Project{
		ID:                1,
		PathWithNamespace: "foo/bar",
		TagList:           []string{"baz", "yolo"},
	}

	expectedValue := ProjectRef{
		Project: Project{
			Name: "foo/bar",
		},
		PathWithNamespace: "foo/bar",
		Kind:              ProjectRefKindTag,
		ID:                1,
		Topics:            "baz,yolo",
		Ref:               "v0.0.7",
		Jobs:              make(map[string]goGitlab.Job),
	}

	assert.Equal(t, expectedValue, NewProjectRef(p, gp, "v0.0.7", ProjectRefKindTag))
}
