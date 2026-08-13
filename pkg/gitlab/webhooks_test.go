package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"
)

func TestGetProjectHooks(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/hooks"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			expectedQueryParams := url.Values{}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `[{"id":1}]`)
		})

	hooks, err := c.GetProjectHooks(ctx, "foo")
	fmt.Println(hooks)
	assert.NoError(t, err)
	assert.Len(t, hooks, 1)
}

func TestAddProjectHook(t *testing.T) {
	ctx, mux, server, c := getMockedClient()
	defer server.Close()

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/hooks"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			expectedQueryParams := url.Values{}
			assert.Equal(t, expectedQueryParams, r.URL.Query())
			fmt.Fprint(w, `{"id":1, "url":"www.example.com/webhook", "push_events":false, "pipeline_events": true,  "deployment_events": true,  "enable_ssl_verification": false}`)
		})

	hook, err := c.AddProjectHook(ctx, "foo", &goGitlab.AddProjectHookOptions{
		PushEvents:            pointy.Bool(false),
		PipelineEvents:        pointy.Bool(true),
		DeploymentEvents:      pointy.Bool(true),
		EnableSSLVerification: pointy.Bool(false), // add config for this later
		URL:                   pointy.String("www.example.com/webhook"),
		Token:                 pointy.String("token"),
	})

	h := goGitlab.ProjectHook{
		URL:                   "www.example.com/webhook",
		ID:                    1,
		PushEvents:            false,
		PipelineEvents:        true,
		DeploymentEvents:      true,
		EnableSSLVerification: false,
	}

	assert.NoError(t, err)
	assert.Equal(t, &h, hook)
}
