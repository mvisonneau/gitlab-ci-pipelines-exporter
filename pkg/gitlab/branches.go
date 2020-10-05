package gitlab

import (
	"regexp"

	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectBranches ..
func (c *Client) GetProjectBranches(projectID int, refsRegexp string) ([]string, error) {
	var names []string

	options := &goGitlab.ListBranchesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	re, err := regexp.Compile(refsRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		branches, resp, err := c.Branches.ListBranches(projectID, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			if re.MatchString(branch.Name) {
				names = append(names, branch.Name)
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}
