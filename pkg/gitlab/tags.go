package gitlab

import (
	"regexp"

	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectTags ..
func (c *Client) GetProjectTags(projectID int, refsRegexp string) ([]string, error) {
	var names []string

	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re, err := regexp.Compile(refsRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		tags, resp, err := c.Tags.ListTags(projectID, options)
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
