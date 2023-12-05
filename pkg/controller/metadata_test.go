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
			expectedVersion: gitlab.NewGitLabVersion("v16.7.0-pre"),
		},
		{
			name: "unsuccessful parse",
			data: `
{
"revision":"3fe364fe754"
}
`,
			expectedVersion: gitlab.NewGitLabVersion(""),
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
