package gitlab

import (
	"regexp"
	"time"

	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectTags ..
func (c *Client) GetProjectTags(projectName string, filterRegexp string) ([]string, error) {
	var names []string

	options := &goGitlab.ListTagsOptions{
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
		tags, resp, err := c.Tags.ListTags(projectName, options)
		if err != nil {
			return names, err
		}

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				names = append(names, tag.Name)
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return names, nil
}

// GetProjectMostRecentTagCommit ..
func (c *Client) GetProjectMostRecentTagCommit(project, filterRegexp string) (string, time.Time, error) {
	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re, err := regexp.Compile(filterRegexp)
	if err != nil {
		return "", time.Time{}, err
	}

	for {
		c.rateLimit()
		tags, resp, err := c.Tags.ListTags(project, options)
		if err != nil {
			return "", time.Time{}, err
		}

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				return tag.Commit.ShortID, *tag.Commit.CommittedDate, nil
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return "", time.Time{}, nil
}
