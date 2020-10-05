package schemas

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"

	goGitlab "github.com/xanzy/go-gitlab"
)

const (
	// ProjectRefKindBranch refers to a branch
	ProjectRefKindBranch ProjectRefKind = "branch"

	// ProjectRefKindTag refers to a tag
	ProjectRefKindTag ProjectRefKind = "tag"

	// ProjectRefKindMergeRequest refers to a tag
	ProjectRefKindMergeRequest ProjectRefKind = "merge-request"
)

// ProjectRefKind is used to determine the kind of the ref
type ProjectRefKind string

// ProjectRef is what we will use a metrics entity on which we will
// perform regular polling operations
type ProjectRef struct {
	Project

	Kind                        ProjectRefKind
	ID                          int
	PathWithNamespace           string
	Topics                      string
	Ref                         string
	MostRecentPipeline          *goGitlab.Pipeline
	MostRecentPipelineVariables string
	Jobs                        map[string]*goGitlab.Job
}

// ProjectRefKey ..
type ProjectRefKey string

// Key ..
func (pr ProjectRef) Key() ProjectRefKey {
	sum := sha1.Sum([]byte(pr.Name + pr.Ref))
	return ProjectRefKey(base64.URLEncoding.EncodeToString(sum[:]))
}

// ProjectsRefs allows us to keep track of all the ProjectRef
// we have configured/discovered
type ProjectsRefs map[ProjectRefKey]ProjectRef

// Count returns the amount of projects refs in the map
func (prs ProjectsRefs) Count() int {
	return len(prs)
}

// DefaultLabelsValues ..
func (pr ProjectRef) DefaultLabelsValues() map[string]string {
	return map[string]string{
		"project":   pr.PathWithNamespace,
		"topics":    pr.Topics,
		"ref":       pr.Ref,
		"kind":      string(pr.Kind),
		"variables": pr.MostRecentPipelineVariables,
	}
}

// NewProjectRef is an helper which returns a new ProjectRef pointer
func NewProjectRef(project Project, gp *goGitlab.Project, ref string, kind ProjectRefKind) *ProjectRef {
	return &ProjectRef{
		Project:           project,
		Kind:              kind,
		ID:                gp.ID,
		PathWithNamespace: gp.PathWithNamespace,
		Topics:            strings.Join(gp.TagList, ","),
		Ref:               ref,
	}
}
