package schemas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestNewPipeline(t *testing.T) {
	updatedAt := time.Date(2020, 10, 01, 13, 05, 10, 0, time.Local)

	gitlabPipeline := goGitlab.Pipeline{
		ID:        20,
		Coverage:  "25.6",
		UpdatedAt: &updatedAt,
		Duration:  15,
		Status:    "pending",
	}

	expectedPipeline := Pipeline{
		ID:              20,
		Coverage:        25.6,
		Timestamp:       1601553910,
		DurationSeconds: 15,
		Status:          "pending",
	}

	assert.Equal(t, expectedPipeline, NewPipeline(gitlabPipeline))
}
