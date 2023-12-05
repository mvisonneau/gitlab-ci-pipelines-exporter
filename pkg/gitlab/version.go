package gitlab

import (
	"strings"

	"golang.org/x/mod/semver"
)

type GitLabVersion struct {
	Version string
}

func NewGitLabVersion(version string) GitLabVersion {
	ver := ""
	if strings.HasPrefix(version, "v") {
		ver = version
	} else if version != "" {
		ver = "v" + version
	}

	return GitLabVersion{Version: ver}
}

// PipelineJobsKeysetPaginationSupported returns true if the GitLab instance
// is running 15.9 or later.
func (v GitLabVersion) PipelineJobsKeysetPaginationSupported() bool {
	if v.Version == "" {
		return false
	}

	return semver.Compare(v.Version, "v15.9.0") >= 0
}
