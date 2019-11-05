package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jpillora/backoff"
	"github.com/xanzy/go-gitlab"

	log "github.com/sirupsen/logrus"
)

func projectExists(p Project) bool {
	for _, cp := range cfg.Projects {
		if cmp.Equal(p, cp) {
			return true
		}
	}
	return false
}

func refExists(refs []string, r string) bool {
	for _, ref := range refs {
		if cmp.Equal(r, ref) {
			return true
		}
	}
	return false
}

func (c *Client) getProject(name string) (*gitlab.Project, error) {
	p, _, err := c.Projects.GetProject(name, &gitlab.GetProjectOptions{})
	return p, err
}

func (c *Client) listProjects(w *Wildcard) ([]Project, error) {
	log.Debugf("Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

	projects := []Project{}
	trueVal := true
	falseVal := false
	listOptions := gitlab.ListOptions{
		PerPage: 20,
		Page:    1,
	}

	for {
		var gps []*gitlab.Project
		var resp *gitlab.Response
		var err error

		switch w.Owner.Kind {
		case "user":
			gps, resp, err = c.Projects.ListUserProjects(
				w.Owner.Name,
				&gitlab.ListProjectsOptions{
					Archived:    &falseVal,
					ListOptions: listOptions,
					Search:      &w.Search,
					Simple:      &trueVal,
				},
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&gitlab.ListGroupProjectsOptions{
					Archived:         &falseVal,
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
					Archived:    &falseVal,
					Simple:      &trueVal,
					Search:      &w.Search,
				},
			)
		}

		if err != nil {
			return projects, fmt.Errorf("Unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
		}

		for _, gp := range gps {
			projects = append(
				projects,
				Project{
					Name: gp.PathWithNamespace,
					Refs: w.Refs,
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

func (c *Client) pollProjectsFromWildcards() {
	for _, w := range cfg.Wildcards {
		foundProjects, err := c.listProjects(&w)
		if err != nil {
			log.Error(err.Error())
		} else {
			for _, p := range foundProjects {
				if !projectExists(p) {
					log.Infof("Found project : %s", p.Name)
					go c.pollProject(p)
					cfg.Projects = append(cfg.Projects, p)
				}
			}
		}
	}
}

func (c *Client) pollRefs(projectID int, refsRegexp string) (refs []*string, err error) {
	if len(refsRegexp) == 0 {
		if len(cfg.DefaultRefsRegexp) > 0 {
			refsRegexp = cfg.DefaultRefsRegexp
		} else {
			refsRegexp = "^master$"
		}
	}

	re := regexp.MustCompile(refsRegexp)

	branches, err := c.pollBranchNames(projectID)
	if err != nil {
		return
	}

	for _, branch := range branches {
		if re.MatchString(*branch) {
			refs = append(refs, branch)
		}
	}

	tags, err := c.pollTagNames(projectID)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if re.MatchString(*tag) {
			refs = append(refs, tag)
		}
	}

	return
}

func (c *Client) pollBranchNames(projectID int) ([]*string, error) {
	var names []*string

	options := &gitlab.ListBranchesOptions{
		PerPage: 20,
		Page:    1,
	}

	for {
		branches, resp, err := c.Branches.ListBranches(projectID, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			names = append(names, &branch.Name)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *Client) pollTagNames(projectID int) ([]*string, error) {
	var names []*string

	options := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		tags, resp, err := c.Tags.ListTags(projectID, options)
		if err != nil {
			return names, err
		}

		for _, tag := range tags {
			names = append(names, &tag.Name)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *Client) pollProject(p Project) {
	var polledRefs []string
	for {
		time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)

		log.Debugf("Fetching project : %s", p.Name)
		gp, err := c.getProject(p.Name)
		if err != nil {
			log.Errorf("Unable to fetch project '%s' from the GitLab API : %v", p.Name, err.Error())
			time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
			continue
		}

		log.Debugf("Polling refs for project : %s", p.Name)

		refs, err := c.pollRefs(gp.ID, p.Refs)
		if err != nil {
			log.Warnf("Could not fetch refs for project '%s'", p.Name)
			continue
		}

		if len(refs) > 0 {
			for _, ref := range refs {
				if !refExists(polledRefs, *ref) {
					log.Infof("Found ref '%s' for project '%s'", *ref, p.Name)
					go c.pollProjectRef(gp, *ref)
					polledRefs = append(polledRefs, *ref)
				}
			}
		} else {
			log.Warnf("No refs found for for project '%s'", p.Name)
		}

	}
}

func (c *Client) pollProjectRef(gp *gitlab.Project, ref string) {
	log.Debugf("Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)
	var lastPipeline *gitlab.Pipeline

	runCount.WithLabelValues(gp.PathWithNamespace, ref).Add(0)

	b := &backoff.Backoff{
		Min:    time.Duration(cfg.PipelinesPollingIntervalSeconds) * time.Second,
		Max:    time.Duration(cfg.PipelinesMaxPollingIntervalSeconds) * time.Second,
		Factor: 1.4,
		Jitter: false,
	}

	for {
		time.Sleep(b.Duration())

		pipelines, _, err := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
		if err != nil {
			log.Errorf("ListProjectPipelines: %s", err.Error())
			continue
		}

		if len(pipelines) == 0 {
			log.Debugf("Could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
		} else if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
			if lastPipeline != nil {
				runCount.WithLabelValues(gp.PathWithNamespace, ref).Inc()
			}

			lastPipeline, _, err = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
			if err != nil {
				log.Errorf("GetPipeline: %s", err.Error())
			}

			if lastPipeline != nil {
				if lastPipeline.Coverage != "" {
					if parsedCoverage, err := strconv.ParseFloat(lastPipeline.Coverage, 64); err == nil {
						coverage.WithLabelValues(gp.PathWithNamespace, ref).Set(parsedCoverage)
					} else {
						log.Warnf("Could not parse coverage string returned from GitLab API: '%s' - '%s'", lastPipeline.Coverage, err.Error())
					}
				}

				lastRunDuration.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(lastPipeline.Duration))
				lastRunID.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(lastPipeline.ID))

				// List of available statuses from the API spec
				// ref: https://docs.gitlab.com/ee/api/pipelines.html#list-project-pipelines
				for _, s := range []string{"running", "pending", "success", "failed", "canceled", "skipped"} {
					if s == lastPipeline.Status {
						lastRunStatus.WithLabelValues(gp.PathWithNamespace, ref, s).Set(1)
					} else {
						lastRunStatus.WithLabelValues(gp.PathWithNamespace, ref, s).Set(0)
					}
				}
			}
		}

		if lastPipeline != nil {
			timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
			b.Reset()
		}
	}
}

func (c *Client) pollProjects() {
	log.Infof("%d project(s) and %d wildcard(s) configured", len(cfg.Projects), len(cfg.Wildcards))
	for _, p := range cfg.Projects {
		go c.pollProject(p)
	}

	for {
		c.pollProjectsFromWildcards()
		time.Sleep(time.Duration(cfg.ProjectsPollingIntervalSeconds) * time.Second)
	}
}
