package controller

import (
	"context"
	"regexp"
	"strconv"

	goGitlab "github.com/xanzy/go-gitlab"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
)

func parseVersionRegex(version string) (int, int, int, string) {
	regexPattern := `^(\d+)\.(\d+)\.(\d+)(?:-(.*))?$`
	r := regexp.MustCompile(regexPattern)

	matches := r.FindStringSubmatch(version)

	if matches == nil {
		return 0, 0, 0, "Invalid version format"
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	suffix := matches[4]

	return major, minor, patch, suffix
}

func (c *Controller) GetGitLabMetadata(ctx context.Context) error {
	options := []goGitlab.RequestOptionFunc{goGitlab.WithContext(ctx)}

	metadata, _, err := c.Gitlab.Metadata.GetMetadata(options...)
	if err != nil {
		return err
	}

	if metadata.Version != "" {
		major, minor, patch, suffix := parseVersionRegex(metadata.Version)
		c.Gitlab.UpdateVersion(
			gitlab.GitLabVersion{
				Major:  major,
				Minor:  minor,
				Patch:  patch,
				Suffix: suffix,
			},
		)
	}

	return nil
}
