package gitlab

import (
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectBranches ..
func (c *Client) GetProjectBranches(projectName, filterRegexp string, maxAgeSeconds uint) ([]string, error) {
	var names []string

	options := &goGitlab.ListBranchesOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re, err := regexp.Compile(filterRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		branches, resp, err := c.Branches.ListBranches(projectName, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			if re.MatchString(branch.Name) {
				if maxAgeSeconds > 0 && time.Now().Sub(*branch.Commit.AuthoredDate) > (time.Duration(maxAgeSeconds)*time.Second) {
					log.WithFields(log.Fields{
						"project-name":    projectName,
						"branch":          branch.Name,
						"regexp":          filterRegexp,
						"max-age-seconds": maxAgeSeconds,
						"authored-date":   *branch.Commit.AuthoredDate,
					}).Debug("branch matching regexp but last authored at a date outside of the required timeframe, ignoring..")
					continue
				}
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
