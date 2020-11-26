package gitlab

import (
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectTags ..
func (c *Client) GetProjectTags(projectName, filterRegexp string, maxAgeSeconds uint) ([]string, error) {
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
				if maxAgeSeconds > 0 && time.Now().Sub(*tag.Commit.AuthoredDate) > (time.Duration(maxAgeSeconds)*time.Second) {
					log.WithFields(log.Fields{
						"project-name":    projectName,
						"tag":             tag.Name,
						"regexp":          filterRegexp,
						"max-age-seconds": maxAgeSeconds,
						"authored-date":   *tag.Commit.AuthoredDate,
					}).Debug("tag matching regexp but last authored at a date outside of the required timeframe, ignoring..")
					continue
				}
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
func (c *Client) GetProjectMostRecentTagCommit(projectName, filterRegexp string) (string, float64, error) {
	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	re, err := regexp.Compile(filterRegexp)
	if err != nil {
		return "", 0, err
	}

	for {
		c.rateLimit()
		tags, resp, err := c.Tags.ListTags(projectName, options)
		if err != nil {
			return "", 0, err
		}

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				return tag.Commit.ShortID, float64(tag.Commit.CommittedDate.Unix()), nil
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return "", 0, nil
}
