package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentKey(t *testing.T) {
	e := Environment{
		ProjectName: "foo",
		Name:        "bar",
	}

	assert.Equal(t, EnvironmentKey("2666930069"), e.Key())
}

func TestEnvironmentsCount(t *testing.T) {
	assert.Equal(t, 2, Environments{
		EnvironmentKey("foo"): Environment{},
		EnvironmentKey("bar"): Environment{},
	}.Count())
}

func TestEnvironmentDefaultLabelsValues(t *testing.T) {
	e := Environment{
		ProjectName: "foo",
		Name:        "bar",
	}

	expectedValue := map[string]string{
		"project":     "foo",
		"environment": "bar",
	}

	assert.Equal(t, expectedValue, e.DefaultLabelsValues())
}

func TestEnvironmentInformationLabelsValues(t *testing.T) {
	e := Environment{
		ProjectName: "foo",
		Name:        "bar",
		ID:          10,
		ExternalURL: "http://genial",
		Available:   true,
		LatestDeployment: Deployment{
			RefKind:       RefKindBranch,
			RefName:       "foo",
			CommitShortID: "123abcde",
			AuthorEmail:   "foo@bar.net",
		},
	}

	expectedValue := map[string]string{
		"project":                 "foo",
		"environment":             "bar",
		"environment_id":          "10",
		"external_url":            "http://genial",
		"kind":                    "branch",
		"ref":                     "foo",
		"current_commit_short_id": "123abcde",
		"latest_commit_short_id":  "",
		"available":               "true",
		"author_email":            "foo@bar.net",
	}

	assert.Equal(t, expectedValue, e.InformationLabelsValues())
}
