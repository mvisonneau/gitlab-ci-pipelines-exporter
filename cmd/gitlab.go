package cmd

import (
	"os"
	"regexp"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func (c *Client) getProject(name string) *gitlab.Project {
	p, _, err := c.Projects.GetProject(name, &gitlab.GetProjectOptions{})
	if err != nil {
		log.Fatalf("Unable to fetch project '%v' from the GitLab API : %v", name, err.Error())
		os.Exit(1)
	}
	return p
}

func (c *Client) pollProjectsFromWildcards() {
	for _, w := range cfg.Wildcards {
		for _, p := range c.listProjects(&w) {
			if !c.projectExists(p) {
				log.Infof("Found project : %s", p.Name)
				go c.pollProject(p)
				cfg.Projects = append(cfg.Projects, p)
			}
		}
	}
}

func (c *Client) projectExists(p Project) bool {
	for _, cp := range cfg.Projects {
		if p == cp {
			return true
		}
	}
	return false
}

func (c *Client) listProjects(w *Wildcard) (projects []Project) {
	log.Infof("Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

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
					ListOptions: listOptions,
					Archived:    &falseVal,
					Simple:      &trueVal,
					Search:      &w.Search,
				},
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&gitlab.ListGroupProjectsOptions{
					ListOptions: listOptions,
					Archived:    &falseVal,
					Simple:      &trueVal,
					Search:      &w.Search,
				},
			)
		default:
			log.Fatalf("Invalid owner kind '%s' must be either 'user' or 'group'", w.Owner.Kind)
			os.Exit(1)
		}

		if err != nil {
			log.Fatalf("Unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
			os.Exit(1)
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

	return
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
		gp := c.getProject(p.Name)
		log.Infof("Polling refs for project : %s", p.Name)

		refs, err := c.pollRefs(gp.ID, p.Refs)
		if err != nil {
			log.Warnf("Could not fetch refs for project '%s'", p.Name)
		}

		if len(refs) == 0 {
			log.Warnf("No refs found for for project '%s'", p.Name)
			return
		}

		for _, ref := range refs {
			if !refExists(polledRefs, *ref) {
				log.Infof("Found ref '%s' for project '%s'", *ref, p.Name)
				go c.pollProjectRef(gp, *ref)
				polledRefs = append(polledRefs, *ref)
			}
		}

		time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
	}
}

func refExists(refs []string, r string) bool {
	for _, ref := range refs {
		if r == ref {
			return true
		}
	}
	return false
}

func (c *Client) pollProjectRef(gp *gitlab.Project, ref string) {
	log.Infof("Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)
	var lastPipeline *gitlab.Pipeline

	runCount.WithLabelValues(gp.PathWithNamespace, ref).Add(0)

	b := &backoff.Backoff{
		Min:    time.Duration(cfg.PipelinesPollingIntervalSeconds) * time.Second,
		Max:    time.Duration(cfg.PipelinesMaxPollingIntervalSeconds) * time.Second,
		Factor: 1.4,
		Jitter: false,
	}

	for {
		pipelines, _, err := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
		if err != nil {
			log.Errorf("ListProjectPipelines: %s", err.Error())
		}

		if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
			if lastPipeline != nil {
				runCount.WithLabelValues(gp.PathWithNamespace, ref).Inc()
			}

			if len(pipelines) > 0 {
				lastPipeline, _, err = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
				if err != nil {
					log.Errorf("GetPipeline: %s", err.Error())
				}

				if lastPipeline != nil {
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
			} else {
				log.Warnf("Could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
			}
		}

		if lastPipeline != nil {
			timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
			b.Reset()
		}

		time.Sleep(b.Duration())
	}
}

func (c *Client) pollProjects() {
	log.Infof("%d project(s) configured", len(cfg.Projects))
	for _, p := range cfg.Projects {
		go c.pollProject(p)
	}

	for {
		c.pollProjectsFromWildcards()
		time.Sleep(time.Duration(cfg.ProjectsPollingIntervalSeconds) * time.Second)
	}
}
