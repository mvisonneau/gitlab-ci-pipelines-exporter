package gitlab

import (
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectBranches ..
func (c *Client) GetProjectBranches(projectID int, search string) ([]string, error) {
	var names []string

	options := &goGitlab.ListBranchesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
		Search: &search,
	}

	for {
		c.rateLimit()
		branches, resp, err := c.Branches.ListBranches(projectID, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			names = append(names, branch.Name)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}
