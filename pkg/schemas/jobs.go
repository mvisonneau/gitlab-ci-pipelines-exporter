package schemas

import (
	goGitlab "github.com/xanzy/go-gitlab"
)

// Job ..
type Job struct {
	ID              int
	Name            string
	Stage           string
	Timestamp       float64
	DurationSeconds float64
	Status          string
	ArtifactSize    float64
	Runner          Runner
}

// Runner ..
type Runner struct {
	Description string
}

// Jobs ..
type Jobs map[string]Job

// NewJob ..
func NewJob(gj goGitlab.Job) Job {
	var artifactSize float64
	for _, artifact := range gj.Artifacts {
		artifactSize += float64(artifact.Size)
	}

	var timestamp float64
	if gj.CreatedAt != nil {
		timestamp = float64(gj.CreatedAt.Unix())
	}

	return Job{
		ID:              gj.ID,
		Name:            gj.Name,
		Stage:           gj.Stage,
		Timestamp:       timestamp,
		DurationSeconds: gj.Duration,
		Status:          gj.Status,
		ArtifactSize:    artifactSize,

		Runner: Runner{
			Description: gj.Runner.Description,
		},
	}
}
