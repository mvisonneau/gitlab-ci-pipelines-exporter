package schemas

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// Pipeline ..
type Pipeline struct {
	ID                    int
	Coverage              float64
	Timestamp             float64
	DurationSeconds       float64
	QueuedDurationSeconds float64
	Status                string
	Variables             string
}

// NewPipeline ..
func NewPipeline(gp goGitlab.Pipeline) Pipeline {
	var (
		coverage  float64
		err       error
		timestamp float64
		queued    time.Duration
	)

	if gp.Coverage != "" {
		coverage, err = strconv.ParseFloat(gp.Coverage, 64)
		if err != nil {
			log.WithField("error", err.Error()).Warnf("could not parse coverage string returned from GitLab API '%s' into Float64", gp.Coverage)
		}
	}

	if gp.UpdatedAt != nil {
		timestamp = float64(gp.UpdatedAt.Unix())
	}

	if gp.StartedAt != nil && gp.CreatedAt != nil {
		if gp.CreatedAt.Before(*gp.StartedAt) {
			queued = gp.StartedAt.Sub(*gp.CreatedAt)
		}
	}

	return Pipeline{
		ID:                    gp.ID,
		Coverage:              coverage,
		Timestamp:             timestamp,
		DurationSeconds:       float64(gp.Duration),
		QueuedDurationSeconds: queued.Seconds(),
		Status:                gp.Status,
	}
}
