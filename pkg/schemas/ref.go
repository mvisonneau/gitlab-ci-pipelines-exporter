package schemas

import (
	"hash/crc32"
	"strconv"
	"strings"

	goGitlab "github.com/xanzy/go-gitlab"
)

const (
	// RefKindBranch refers to a branch
	RefKindBranch RefKind = "branch"

	// RefKindTag refers to a tag
	RefKindTag RefKind = "tag"

	// RefKindMergeRequest refers to a tag
	RefKindMergeRequest RefKind = "merge-request"
)

// RefKind is used to determine the kind of the ref
type RefKind string

// Ref is what we will use a metrics entity on which we will
// perform regular pulling operations
type Ref struct {
	Project

	Kind                        RefKind
	ID                          int
	PathWithNamespace           string
	Topics                      string
	Ref                         string
	MostRecentPipeline          *goGitlab.Pipeline
	MostRecentPipelineVariables string
	Jobs                        map[string]goGitlab.Job
}

// RefKey ..
type RefKey string

// Key ..
func (ref Ref) Key() RefKey {
	return RefKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(ref.PathWithNamespace + ref.Ref)))))
}

// Refs allows us to keep track of all the Ref
// we have configured/discovered
type Refs map[RefKey]Ref

// Count returns the amount of projects refs in the map
func (refs Refs) Count() int {
	return len(refs)
}

// DefaultLabelsValues ..
func (ref Ref) DefaultLabelsValues() map[string]string {
	return map[string]string{
		"project":   ref.PathWithNamespace,
		"topics":    ref.Topics,
		"ref":       ref.Ref,
		"kind":      string(ref.Kind),
		"variables": ref.MostRecentPipelineVariables,
	}
}

// NewRef is an helper which returns a new Ref pointer
func NewRef(project Project, gp *goGitlab.Project, ref string, kind RefKind) Ref {
	return Ref{
		Project:           project,
		Kind:              kind,
		ID:                gp.ID,
		PathWithNamespace: gp.PathWithNamespace,
		Topics:            strings.Join(gp.TagList, ","),
		Ref:               ref,
		Jobs:              make(map[string]goGitlab.Job),
	}
}
