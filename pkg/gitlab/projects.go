package gitlab

import (
	"context"
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.openly.dev/pointy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// GetProject ..
func (c *Client) GetProject(ctx context.Context, name string) (*goGitlab.Project, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:GetProject")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", name))

	log.WithFields(log.Fields{
		"project-name": name,
	}).Debug("reading project")

	c.rateLimit(ctx)
	p, resp, err := c.Projects.GetProject(name, &goGitlab.GetProjectOptions{}, goGitlab.WithContext(ctx))
	c.requestsRemaining(resp)

	return p, err
}

// ListProjects ..
func (c *Client) ListProjects(ctx context.Context, w config.Wildcard) ([]schemas.Project, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListProjects")
	defer span.End()
	span.SetAttributes(attribute.String("wildcard_search", w.Search))
	span.SetAttributes(attribute.String("wildcard_owner_kind", w.Owner.Kind))
	span.SetAttributes(attribute.String("wildcard_owner_name", w.Owner.Name))
	span.SetAttributes(attribute.Bool("wildcard_owner_include_subgroups", w.Owner.IncludeSubgroups))
	span.SetAttributes(attribute.Bool("wildcard_archived", w.Archived))

	logFields := log.Fields{
		"wildcard-search":                  w.Search,
		"wildcard-owner-kind":              w.Owner.Kind,
		"wildcard-owner-name":              w.Owner.Name,
		"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
		"wildcard-archived":                w.Archived,
	}
	log.WithFields(logFields).Debug("listing all projects from wildcard")

	var projects []schemas.Project

	listOptions := goGitlab.ListOptions{
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
		var (
			gps  []*goGitlab.Project
			resp *goGitlab.Response
			err  error
		)

		c.rateLimit(ctx)

		switch w.Owner.Kind {
		case "user":
			gps, resp, err = c.Projects.ListUserProjects(
				w.Owner.Name,
				&goGitlab.ListProjectsOptions{
					Archived:    &w.Archived,
					ListOptions: listOptions,
					Search:      &w.Search,
					Simple:      pointy.Bool(true),
				},
				goGitlab.WithContext(ctx),
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&goGitlab.ListGroupProjectsOptions{
					Archived:         &w.Archived,
					WithShared:       pointy.Bool(false),
					IncludeSubGroups: &w.Owner.IncludeSubgroups,
					ListOptions:      listOptions,
					Search:           &w.Search,
					Simple:           pointy.Bool(true),
				},
				goGitlab.WithContext(ctx),
			)
		default:
			// List all visible projects
			gps, resp, err = c.Projects.ListProjects(
				&goGitlab.ListProjectsOptions{
					ListOptions: listOptions,
					Archived:    &w.Archived,
					Search:      &w.Search,
					Simple:      pointy.Bool(true),
				},
				goGitlab.WithContext(ctx),
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
