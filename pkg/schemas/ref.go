package schemas

import (
	"hash/crc32"
	"strconv"
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
	Kind           RefKind
	ProjectName    string
	Name           string
	Topics         string
	LatestPipeline Pipeline
	LatestJobs     Jobs

	OutputSparseStatusMetrics                          bool
	PullPipelineJobsEnabled                            bool
	PullPipelineJobsFromChildPipelinesEnabled          bool
	PullPipelineJobsRunnerDescriptionEnabled           bool
	PullPipelineJobsRunnerDescriptionAggregationRegexp string
	PullPipelineJobsTraceRules                         []string
	PullPipelineVariablesEnabled                       bool
	PullPipelineVariablesRegexp                        string
}

// RefKey ..
type RefKey string

// Key ..
func (ref Ref) Key() RefKey {
	return RefKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(string(ref.Kind) + ref.ProjectName + ref.Name)))))
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
		"kind":      string(ref.Kind),
		"project":   ref.ProjectName,
		"ref":       ref.Name,
		"topics":    ref.Topics,
		"variables": ref.LatestPipeline.Variables,
	}
}

// NewRef is an helper which returns a new Ref pointer
func NewRef(
	kind RefKind,
	projectName, name, topics string,
	outputSparseStatusMetrics, pullPipelineJobsEnabled, pullPipelineJobsFromChildPipelinesEnabled, pullPipelineJobsRunnerDescriptionEnabled, pullPipelineVariablesEnabled bool,
	pullPipelineVariablesRegexp, pullPipelineJobsRunnerDescriptionAggregationRegexp string, pullPipelineJobsTraceRules []string,
) Ref {
	return Ref{
		Kind:        kind,
		ProjectName: projectName,
		Name:        name,
		Topics:      topics,
		LatestJobs:  make(Jobs),

		OutputSparseStatusMetrics:                          outputSparseStatusMetrics,
		PullPipelineJobsEnabled:                            pullPipelineJobsEnabled,
		PullPipelineJobsFromChildPipelinesEnabled:          pullPipelineJobsFromChildPipelinesEnabled,
		PullPipelineJobsRunnerDescriptionEnabled:           pullPipelineJobsRunnerDescriptionEnabled,
		PullPipelineJobsRunnerDescriptionAggregationRegexp: pullPipelineJobsRunnerDescriptionAggregationRegexp,
		PullPipelineJobsTraceRules:                         pullPipelineJobsTraceRules,
		PullPipelineVariablesEnabled:                       pullPipelineVariablesEnabled,
		PullPipelineVariablesRegexp:                        pullPipelineVariablesRegexp,
	}
}
