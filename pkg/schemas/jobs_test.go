package schemas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestNewJob(t *testing.T) {
	createdAt := time.Date(2020, 10, 01, 13, 05, 05, 0, time.UTC)
	gitlabJob := goGitlab.Job{
		ID:        2,
		Name:      "foo",
		CreatedAt: &createdAt,
		Duration:  15,
		Status:    "failed",
		Stage:     "🚀",
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
		ID:              2,
		Name:            "foo",
		Stage:           "🚀",
		Timestamp:       1.601557505e+09,
		DurationSeconds: 15,
		Status:          "failed",
		ArtifactSize:    150,
	}

	assert.Equal(t, expectedJob, NewJob(gitlabJob))
}
