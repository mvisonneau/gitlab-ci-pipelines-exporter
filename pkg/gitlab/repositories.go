package gitlab

import (
	"github.com/openlyinc/pointy"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetCommitCountBetweenRefs ..
func (c *Client) GetCommitCountBetweenRefs(project, from, to string) (int, error) {
	log.WithFields(log.Fields{
		"project-name": project,
		"from-ref":     from,
		"to-ref":       to,
	}).Debug("comparing refs")

	c.rateLimit()
	cmp, _, err := c.Repositories.Compare(project, &goGitlab.CompareOptions{
		From:     &from,
		To:       &to,
		Straight: pointy.Bool(true),
	}, nil)
	return len(cmp.Commits), err
}
