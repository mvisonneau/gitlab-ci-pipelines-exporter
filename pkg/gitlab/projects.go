package gitlab

import (
	"fmt"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProject ..
func (c *Client) GetProject(name string) (*goGitlab.Project, error) {
	c.rateLimit()
	log.WithFields(
		log.Fields{
			"project-name": name,
		},
	).Debug("reading project")
	p, _, err := c.Projects.GetProject(name, &goGitlab.GetProjectOptions{})
	return p, err
}

// ListProjects ..
func (c *Client) ListProjects(w schemas.Wildcard) ([]schemas.Project, error) {
	log.WithFields(
		log.Fields{
			"wildcard-search":                  w.Search,
			"wildcard-owner-kind":              w.Owner.Kind,
			"wildcard-owner-name":              w.Owner.Name,
			"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
			"wildcard-archived":                w.Archived,
		},
	).Debug("listing all projects from wildcard")

	var projects []schemas.Project
	trueVal := true
	listOptions := gitlab.ListOptions{
		PerPage: 20,
		Page:    1,
	}

	for {
		var gps []*gitlab.Project
		var resp *gitlab.Response
		var err error

		c.rateLimit()
		switch w.Owner.Kind {
		case "user":
			gps, resp, err = c.Projects.ListUserProjects(
				w.Owner.Name,
				&gitlab.ListProjectsOptions{
					Archived:    &w.Archived,
					ListOptions: listOptions,
					Search:      &w.Search,
					Simple:      &trueVal,
				},
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&gitlab.ListGroupProjectsOptions{
					Archived:         &w.Archived,
					IncludeSubgroups: &w.Owner.IncludeSubgroups,
					ListOptions:      listOptions,
					Search:           &w.Search,
					Simple:           &trueVal,
				},
			)
		default:
			gps, resp, err = c.Projects.ListProjects(
				&gitlab.ListProjectsOptions{
					ListOptions: listOptions,
					Archived:    &w.Archived,
					Simple:      &trueVal,
					Search:      &w.Search,
				},
			)
		}

		if err != nil {
			return projects, fmt.Errorf("unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
		}

		// Copy relevant settings from wildcard into created project
		for _, gp := range gps {
			projects = append(
				projects,
				schemas.Project{
					Parameters: w.Parameters,
					Name:       gp.PathWithNamespace,
				},
			)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		listOptions.Page = resp.NextPage
	}

	return projects, nil
}
