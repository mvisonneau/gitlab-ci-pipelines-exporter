package schemas

import (
	"strconv"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// Pipeline ..
type Pipeline struct {
	ID              int
	Coverage        float64
	Timestamp       float64
	DurationSeconds float64
	Status          string
	Variables       string
}

// NewPipeline ..
func NewPipeline(gp goGitlab.Pipeline) Pipeline {
	var coverage float64
	var err error
	if gp.Coverage != "" {
		coverage, err = strconv.ParseFloat(gp.Coverage, 64)
		if err != nil {
			log.WithField("error", err.Error()).Warnf("could not parse coverage string returned from GitLab API '%s' into Float64", gp.Coverage)
		}
	}

	var timestamp float64
	if gp.UpdatedAt != nil {
		timestamp = float64(gp.UpdatedAt.Unix())
	}

	return Pipeline{
		ID:              gp.ID,
		Coverage:        coverage,
		Timestamp:       timestamp,
		DurationSeconds: float64(gp.Duration),
		Status:          gp.Status,
	}
}
