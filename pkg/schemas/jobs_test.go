package schemas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestNewJob(t *testing.T) {
	createdAt := time.Date(2020, 10, 1, 13, 5, 5, 0, time.UTC)
	startedAt := time.Date(2020, 10, 1, 13, 5, 35, 0, time.UTC)

	gitlabJob := goGitlab.Job{
		ID:             2,
		Name:           "foo",
		CreatedAt:      &createdAt,
		StartedAt:      &startedAt,
		Duration:       15,
		QueuedDuration: 10,
		Status:         "failed",
		Stage:          "🚀",
		TagList:        []string{"test-tag"},
		Runner: struct {
			ID          int    "json:\"id\""
			Description string "json:\"description\""
			Active      bool   "json:\"active\""
			IsShared    bool   "json:\"is_shared\""
			Name        string "json:\"name\""
		}{
			Description: "xxx",
		},
		Artifacts: []struct {
			FileType   string "json:\"file_type\""
			Filename   string "json:\"filename\""
			Size       int    "json:\"size\""
			FileFormat string "json:\"file_format\""
		}{
			{
				Size: 100,
			},
			{
				Size: 50,
			},
		},
	}

	expectedJob := Job{
		ID:                    2,
		Name:                  "foo",
		Stage:                 "🚀",
		Timestamp:             1.601557505e+09,
		DurationSeconds:       15,
		QueuedDurationSeconds: 10,
		Status:                "failed",
		TagList:               "test-tag",
		ArtifactSize:          150,

		Runner: Runner{
			Description: "xxx",
		},
	}

	assert.Equal(t, expectedJob, NewJob(gitlabJob))
}
