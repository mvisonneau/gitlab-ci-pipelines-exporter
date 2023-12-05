package controller

import (
	"context"

	goGitlab "github.com/xanzy/go-gitlab"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
)

func (c *Controller) GetGitLabMetadata(ctx context.Context) error {
	options := []goGitlab.RequestOptionFunc{goGitlab.WithContext(ctx)}

	metadata, _, err := c.Gitlab.Metadata.GetMetadata(options...)
	if err != nil {
		return err
	}

	if metadata.Version != "" {
		c.Gitlab.UpdateVersion(gitlab.NewGitLabVersion(metadata.Version))
	}

	return nil
}
