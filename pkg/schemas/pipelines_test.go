package schemas

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestNewPipeline(t *testing.T) {
	createdAt := time.Date(2020, 10, 1, 13, 4, 10, 0, time.UTC)
	startedAt := time.Date(2020, 10, 1, 13, 5, 10, 0, time.UTC)
	updatedAt := time.Date(2020, 10, 1, 13, 5, 50, 0, time.UTC)

	gitlabPipeline := goGitlab.Pipeline{
		ID:             21,
		Coverage:       "25.6",
		CreatedAt:      &createdAt,
		StartedAt:      &startedAt,
		UpdatedAt:      &updatedAt,
		Duration:       15,
		QueuedDuration: 5,
		Status:         "running",
	}

	expectedPipeline := Pipeline{
		ID:                    21,
		Coverage:              25.6,
		Timestamp:             1.60155755e+09,
		DurationSeconds:       15,
		QueuedDurationSeconds: 5,
		Status:                "running",
	}

	assert.Equal(t, expectedPipeline, NewPipeline(context.Background(), gitlabPipeline))
}

func TestNewTestReport(t *testing.T) {
	gitlabTestReport := goGitlab.PipelineTestReport{
		TotalTime:    10,
		TotalCount:   2,
		SuccessCount: 1,
		FailedCount:  1,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []*goGitlab.PipelineTestSuites{
			{
				Name:         "First",
				TotalTime:    3,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []*goGitlab.PipelineTestCases{
					{
						Name:          "First",
						Classname:     "ClassFirst",
						ExecutionTime: 4,
						Status:        "success",
					},
				},
			},
			{
				Name:         "Second",
				TotalTime:    2,
				TotalCount:   1,
				SuccessCount: 0,
				FailedCount:  1,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []*goGitlab.PipelineTestCases{
					{
						Name:          "First",
						Classname:     "ClassFirst",
						ExecutionTime: 4,
						Status:        "success",
					},
				},
			},
		},
	}

	expectedTestReport := TestReport{
		TotalTime:    10,
		TotalCount:   2,
		SuccessCount: 1,
		FailedCount:  1,
		SkippedCount: 0,
		ErrorCount:   0,
		TestSuites: []TestSuite{
			{
				Name:         "First",
				TotalTime:    3,
				TotalCount:   1,
				SuccessCount: 1,
				FailedCount:  0,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []TestCase{
					{
						Name:          "First",
						Classname:     "ClassFirst",
						ExecutionTime: 4,
						Status:        "success",
					},
				},
			},
			{
				Name:         "Second",
				TotalTime:    2,
				TotalCount:   1,
				SuccessCount: 0,
				FailedCount:  1,
				SkippedCount: 0,
				ErrorCount:   0,
				TestCases: []TestCase{
					{
						Name:          "First",
						Classname:     "ClassFirst",
						ExecutionTime: 4,
						Status:        "success",
					},
				},
			},
		},
	}
	assert.Equal(t, expectedTestReport, NewTestReport(gitlabTestReport))
}

func TestNewTestSuite(t *testing.T) {
	gitlabTestSuite := &goGitlab.PipelineTestSuites{
		Name:         "Suite",
		TotalTime:    4,
		TotalCount:   6,
		SuccessCount: 2,
		FailedCount:  2,
		SkippedCount: 1,
		ErrorCount:   1,
		TestCases: []*goGitlab.PipelineTestCases{
			{
				Name:          "First",
				Classname:     "ClassFirst",
				ExecutionTime: 4,
				Status:        "success",
			},
		},
	}

	expectedTestSuite := TestSuite{
		Name:         "Suite",
		TotalTime:    4,
		TotalCount:   6,
		SuccessCount: 2,
		FailedCount:  2,
		SkippedCount: 1,
		ErrorCount:   1,
		TestCases: []TestCase{
			{
				Name:          "First",
				Classname:     "ClassFirst",
				ExecutionTime: 4,
				Status:        "success",
			},
		},
	}
	assert.Equal(t, expectedTestSuite, NewTestSuite(gitlabTestSuite))
}
