package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
)

func TestGetGitLabMetadataSuccess(t *testing.T) {
	tests := []struct {
		name            string
		data            string
		expectedVersion gitlab.GitLabVersion
	}{
		{
			name: "successful parse",
			data: `
{
"version":"16.7.0-pre",
"revision":"3fe364fe754",
"kas":{
	"enabled":true,
	"externalUrl":"wss://kas.gitlab.com",
	"version":"v16.7.0-rc2"
},
"enterprise":true
}
`,
			expectedVersion: gitlab.GitLabVersion{Major: 16, Minor: 7, Patch: 0, Suffix: "pre"},
		},
		{
			name: "unsuccessful parse",
			data: `
{
"version":"16.7"
}
`,
			expectedVersion: gitlab.GitLabVersion{Major: 0, Minor: 0, Patch: 0, Suffix: "Invalid version format"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, c, mux, srv := newTestController(config.Config{})
			defer srv.Close()

			mux.HandleFunc("/api/v4/metadata",
				func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, tc.data)
				})

			assert.NoError(t, c.GetGitLabMetadata(ctx))
			assert.Equal(t, tc.expectedVersion, c.Gitlab.Version())
		})
	}
}
