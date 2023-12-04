package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineJobsKeysetPaginationSupported(t *testing.T) {
	tests := []struct {
		name           string
		version        GitLabVersion
		expectedResult bool
	}{
		{
			name:           "unknown",
			version:        NewGitLabVersion(""),
			expectedResult: false,
		},
		{
			name:           "15.8.0",
			version:        NewGitLabVersion("15.8.0"),
			expectedResult: false,
		},
		{
			name:           "v15.8.0",
			version:        NewGitLabVersion("v15.8.0"),
			expectedResult: false,
		},
		{
			name:           "15.9.0",
			version:        NewGitLabVersion("15.9.0"),
			expectedResult: true,
		},
		{
			name:           "v15.9.0",
			version:        NewGitLabVersion("v15.9.0"),
			expectedResult: true,
		},
		{
			name:           "15.9.1",
			version:        NewGitLabVersion("15.9.1"),
			expectedResult: true,
		},
		{
			name:           "15.10.2",
			version:        NewGitLabVersion("15.10.2"),
			expectedResult: true,
		},
		{
			name:           "16.0.0",
			version:        NewGitLabVersion("16.0.0"),
			expectedResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.version.PipelineJobsKeysetPaginationSupported()

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
