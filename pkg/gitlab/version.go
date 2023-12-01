package gitlab

type GitLabVersion struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

// PipelineJobsKeysetPaginationSupported returns true if the GitLab instance
// is running 15.9 or later.
func (v GitLabVersion) PipelineJobsKeysetPaginationSupported() bool {
	if v.Major == 0 {
		return false
	} else if v.Major < 15 {
		return false
	} else if v.Major > 15 {
		return true
	} else {
		return v.Minor >= 9
	}
}
