package gitlab

import (
	"fmt"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	goGitlab "github.com/xanzy/go-gitlab"
)

// GetProject ..
func (c *Client) GetProject(name string) (*goGitlab.Project, error) {
	log.WithFields(log.Fields{
		"project-name": name,
	}).Debug("reading project")

	c.rateLimit()
	p, resp, err := c.Projects.GetProject(name, &goGitlab.GetProjectOptions{})
	c.requestsRemaining(resp)

	return p, err
}

// ListProjects ..
func (c *Client) ListProjects(w config.Wildcard) ([]schemas.Project, error) {
	logFields := log.Fields{
		"wildcard-search":                  w.Search,
		"wildcard-owner-kind":              w.Owner.Kind,
		"wildcard-owner-name":              w.Owner.Name,
		"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
		"wildcard-archived":                w.Archived,
	}
	log.WithFields(logFields).Debug("listing all projects from wildcard")

	var projects []schemas.Project
	listOptions := gitlab.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	// As a result, the API will return the projects that the owner has access onto.
	// This is not necessarily what we the end-user would intend when leveraging a
	// scoped wildcard. Therefore, if the wildcard owner name is set, we want to filter
	// out to project actually *belonging* to the owner.
	var ownerRegexp *regexp.Regexp
	if len(w.Owner.Name) > 0 {
		ownerRegexp = regexp.MustCompile(fmt.Sprintf(`^%s\/`, w.Owner.Name))
	} else {
		ownerRegexp = regexp.MustCompile(`.*`)
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
				},
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&gitlab.ListGroupProjectsOptions{
					Archived:         &w.Archived,
					WithShared:       pointy.Bool(false),
					IncludeSubGroups: &w.Owner.IncludeSubgroups,
					ListOptions:      listOptions,
					Search:           &w.Search,
				},
			)
		default:
			// List all visible projects
			gps, resp, err = c.Projects.ListProjects(
				&gitlab.ListProjectsOptions{
					ListOptions: listOptions,
					Archived:    &w.Archived,
					Search:      &w.Search,
				},
			)
		}

		if err != nil {
			return projects, fmt.Errorf("unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
		}
		c.requestsRemaining(resp)

		// Copy relevant settings from wildcard into created project
		for _, gp := range gps {
			if !ownerRegexp.MatchString(gp.PathWithNamespace) {
				log.WithFields(logFields).WithFields(log.Fields{
					"project-id":   gp.ID,
					"project-name": gp.PathWithNamespace,
				}).Debug("project path not matching owner's name, skipping")
				continue
			}

			if !gp.JobsEnabled {
				log.WithFields(logFields).WithFields(log.Fields{
					"project-id":   gp.ID,
					"project-name": gp.PathWithNamespace,
				}).Debug("jobs/pipelines not enabled on project, skipping")
				continue
			}

			p := schemas.NewProject(gp.PathWithNamespace)
			p.ProjectParameters = w.ProjectParameters
			projects = append(projects, p)
		}

		if resp.CurrentPage >= resp.NextPage {
			break
		}

		listOptions.Page = resp.NextPage
	}

	return projects, nil
}
