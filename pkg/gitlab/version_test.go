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
			version:        GitLabVersion{Major: 0, Minor: 0, Patch: 0, Suffix: ""},
			expectedResult: false,
		},
		{
			name:           "15.8.0",
			version:        GitLabVersion{Major: 15, Minor: 8, Patch: 0, Suffix: ""},
			expectedResult: false,
		},
		{
			name:           "15.9.0",
			version:        GitLabVersion{Major: 15, Minor: 9, Patch: 0, Suffix: ""},
			expectedResult: true,
		},
		{
			name:           "15.9.1",
			version:        GitLabVersion{Major: 15, Minor: 9, Patch: 1, Suffix: ""},
			expectedResult: true,
		},
		{
			name:           "15.10.2",
			version:        GitLabVersion{Major: 15, Minor: 10, Patch: 12, Suffix: ""},
			expectedResult: true,
		},
		{
			name:           "16.0.0",
			version:        GitLabVersion{Major: 16, Minor: 0, Patch: 0, Suffix: ""},
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
