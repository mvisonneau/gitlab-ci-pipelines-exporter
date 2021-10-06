package gitlab

import (
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectBranches ..
func (c *Client) GetProjectBranches(p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	refs = make(schemas.Refs)

	options := &goGitlab.ListBranchesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var re *regexp.Regexp
	if re, err = regexp.Compile(p.Pull.Refs.Branches.Regexp); err != nil {
		return
	}

	for {
		c.rateLimit()
		var branches []*goGitlab.Branch
		var resp *goGitlab.Response
		branches, resp, err = c.Branches.ListBranches(p.Name, options)
		if err != nil {
			return
		}

		for _, branch := range branches {
			if re.MatchString(branch.Name) {
				ref := schemas.NewRef(p, schemas.RefKindBranch, branch.Name)
				refs[ref.Key()] = ref
			}
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}
		options.Page = resp.NextPage
	}

	return
}

// GetBranchLatestCommit ..
func (c *Client) GetBranchLatestCommit(project, branch string) (string, float64, error) {
	log.WithFields(log.Fields{
		"project-name": project,
		"branch":       branch,
	}).Debug("reading project branch")

	c.rateLimit()
	b, _, err := c.Branches.GetBranch(project, branch, nil)
	if err != nil {
		return "", 0, err
	}

	return b.Commit.ShortID, float64(b.Commit.CommittedDate.Unix()), nil
}
