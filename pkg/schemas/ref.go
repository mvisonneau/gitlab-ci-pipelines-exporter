package schemas

import (
	"fmt"
	"hash/crc32"
	"regexp"
	"strconv"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
)

const (
	mergeRequestRegexp string = `^((\d+)|refs/merge-requests/(\d+)/head)$`

	// RefKindBranch refers to a branch.
	RefKindBranch RefKind = "branch"

	// RefKindTag refers to a tag.
	RefKindTag RefKind = "tag"

	// RefKindMergeRequest refers to a tag.
	RefKindMergeRequest RefKind = "merge-request"
)

// RefKind is used to determine the kind of the ref.
type RefKind string

// Ref is what we will use a metrics entity on which we will
// perform regular pulling operations.
type Ref struct {
	Kind           RefKind
	Name           string
	Project        Project
	LatestPipeline Pipeline
	LatestJobs     Jobs
}

// RefKey ..
type RefKey string

// Key ..
func (ref Ref) Key() RefKey {
	return RefKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(string(ref.Kind) + ref.Project.Name + ref.Name)))))
}

// Refs allows us to keep track of all the Ref
// we have configured/discovered.
type Refs map[RefKey]Ref

// Count returns the amount of projects refs in the map.
func (refs Refs) Count() int {
	return len(refs)
}

// DefaultLabelsValues ..
func (ref Ref) DefaultLabelsValues() map[string]string {
	return map[string]string{
		"kind":      string(ref.Kind),
		"project":   ref.Project.Name,
		"ref":       ref.Name,
		"topics":    ref.Project.Topics,
		"variables": ref.LatestPipeline.Variables,
		"source":    ref.LatestPipeline.Source,
	}
}

// NewRef is an helper which returns a new Ref.
func NewRef(
	project Project,
	kind RefKind,
	name string,
) Ref {
	return Ref{
		Kind:       kind,
		Name:       name,
		Project:    project,
		LatestJobs: make(Jobs),
	}
}

// GetRefRegexp returns the expected regexp given a ProjectPullRefs config and a RefKind.
func GetRefRegexp(ppr config.ProjectPullRefs, rk RefKind) (re *regexp.Regexp, err error) {
	switch rk {
	case RefKindBranch:
		return regexp.Compile(ppr.Branches.Regexp)
	case RefKindTag:
		return regexp.Compile(ppr.Tags.Regexp)
	case RefKindMergeRequest:
		return regexp.Compile(mergeRequestRegexp)
	}

	return nil, fmt.Errorf("invalid ref kind (%v)", rk)
}

// GetMergeRequestIIDFromRefName parse a refName to extract a merge request IID.
func GetMergeRequestIIDFromRefName(refName string) (string, error) {
	re := regexp.MustCompile(mergeRequestRegexp)
	if matches := re.FindStringSubmatch(refName); len(matches) == 4 {
		if len(matches[2]) > 0 {
			return matches[2], nil
		}

		if len(matches[3]) > 0 {
			return matches[3], nil
		}
	}

	return refName, fmt.Errorf("unable to extract the merge-request ID from the ref (%s)", refName)
}
