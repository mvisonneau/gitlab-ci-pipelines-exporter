package gitlab

import (
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProjectTags ..
func (c *Client) GetProjectTags(p schemas.Project) (
	refs schemas.Refs,
	err error,
) {
	refs = make(schemas.Refs)

	options := &goGitlab.ListTagsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var re *regexp.Regexp
	if re, err = regexp.Compile(p.Pull.Refs.Tags.Regexp); err != nil {
		return
	}

	for {
		c.rateLimit()
		var tags []*goGitlab.Tag
		var resp *goGitlab.Response
		tags, resp, err = c.Tags.ListTags(p.Name, options)
		if err != nil {
			return
		}

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				ref := schemas.NewRef(p, schemas.RefKindTag, tag.Name)
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

		if resp.CurrentPage >= resp.NextPage {
			break
		}
		options.Page = resp.NextPage
	}

	return "", 0, nil
}
