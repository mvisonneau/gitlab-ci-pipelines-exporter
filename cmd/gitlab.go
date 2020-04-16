package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jpillora/backoff"
	"github.com/prometheus/client_golang/prometheus"
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
	c.rateLimit()
	p, _, err := c.Projects.GetProject(name, &gitlab.GetProjectOptions{})
	return p, err
}

func (c *Client) listProjects(w *Wildcard) ([]Project, error) {
	log.Debugf("Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

	projects := []Project{}
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
			return projects, fmt.Errorf("Unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
		}

		// Copy relevant settings from wildcard into created project
		for _, gp := range gps {

			projects = append(
				projects,
				Project{
					Name:                      gp.PathWithNamespace,
					Refs:                      w.Refs,
					FetchPipelineJobMetrics:   w.FetchPipelineJobMetrics,
					OutputSparseStatusMetrics: w.OutputSparseStatusMetrics,
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

func (c *Client) pollRefs(projectID int, refsRegexp string) (refs []string, err error) {
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
			refs = append(refs, *branch)
		}
	}

	tags, err := c.pollTagNames(projectID)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if re.MatchString(*tag) {
			refs = append(refs, *tag)
		}
	}

	return
}

func (c *Client) pollBranchNames(projectID int) ([]*string, error) {
	var names []*string

	options := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	for {
		c.rateLimit()
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
		c.rateLimit()
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
	var err error

	for {
		log.Debugf("Fetching project : %s", p.Name)
		p.GitlabProject, err = c.getProject(p.Name)
		if err != nil {
			log.Errorf("Unable to fetch project '%s' from the GitLab API : %v", p.Name, err.Error())
			time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
			continue
		}

		if cfg.OnInitFetchRefsFromPipelines {
			log.Debugf("Polling project refs %s from most recent %d pipelines (init only)", p.Name, cfg.OnInitFetchRefsFromPipelinesDepthLimit)
			refs, err := c.pollProjectRefsFromPipelines(p.GitlabProject.ID, cfg.OnInitFetchRefsFromPipelinesDepthLimit)
			if err != nil {
				log.Errorf("Unable to fetch refs from project pipelines %s : %v", p.Name, err.Error())
				time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
				continue
			}

			log.Debugf("Found %d refs from %s pipelines", len(refs), p.Name)
			for _, r := range refs {
				log.Infof("Found ref '%s' for project '%s'", r, p.Name)
				go c.pollProjectRef(p, r)
				polledRefs = append(polledRefs, r)
			}
		}

		break
	}

	for {
		log.Debugf("Polling refs for project : %s", p.Name)

		refs, err := c.pollRefs(p.GitlabProject.ID, p.Refs)
		if err != nil {
			log.Warnf("Could not fetch refs for project '%s'", p.Name)
			continue
		}

		if len(refs) > 0 {
			for _, r := range refs {
				if !refExists(polledRefs, r) {
					log.Infof("Found ref '%s' for project '%s'", r, p.Name)
					go c.pollProjectRef(p, r)
					polledRefs = append(polledRefs, r)
				}
			}
		} else {
			log.Warnf("No refs found for project '%s'", p.Name)
		}

		time.Sleep(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
	}
}

func outputStatusMetric(metric *prometheus.GaugeVec, labels []string, statuses []string, status string, sparseMetrics bool) {
	// Moved into separate function to reduce cyclomatic complexity
	// List of available statuses from the API spec
	// ref: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
	for _, s := range statuses {
		args := append(labels, s)
		if s == status {
			metric.WithLabelValues(args...).Set(1)
		} else {
			if sparseMetrics {
				metric.DeleteLabelValues(args...)
			} else {
				metric.WithLabelValues(args...).Set(0)
			}
		}
	}
}

func (c *Client) outputPipelineStatusMetric(status string, sparseMetrics bool, labels ...string) {
	outputStatusMetric(
		lastRunStatus,
		labels,
		[]string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"},
		status,
		sparseMetrics,
	)
}
func (c *Client) outputJobStatusMetric(status string, sparseMetrics bool, labels ...string) {
	outputStatusMetric(
		lastRunJobStatus,
		labels,
		[]string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"},
		status,
		sparseMetrics,
	)
}

func (c *Client) pollPipelineJobs(p Project, pipelineID int, topics string, ref string) error {
	var jobs []*gitlab.Job
	var resp *gitlab.Response
	var gp *gitlab.Project = p.GitlabProject
	var err error

	options := &gitlab.ListJobsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		c.rateLimit()

		jobs, resp, err = c.Jobs.ListPipelineJobs(gp.ID, pipelineID, options)
		if err != nil {
			return err
		}

		jobCount := len(jobs)

		log.Infof("Found %d jobs for pipeline %d", jobCount, pipelineID)

		if jobCount > 0 {
			for _, job := range jobs {
				jobName := job.Name
				stageName := job.Stage
				log.Debugf("Job %s for pipeline %d", jobName, pipelineID)
				lastRunJobDuration.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(float64(job.Duration))

				c.outputJobStatusMetric(job.Status, p.ShouldOutputSparseStatusMetrics(cfg), gp.PathWithNamespace, topics, ref, stageName, jobName)

				timeSinceLastJobRun.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(float64(time.Since(*job.CreatedAt).Round(time.Second).Seconds()))

				jobRunCount.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Inc()

				artifactSize := 0
				for _, artifact := range job.Artifacts {
					artifactSize += artifact.Size
				}

				lastRunJobArtifactSize.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(float64(artifactSize))
			}

		} else {
			log.Warnf("No jobs found for pipeline %d", pipelineID)
			break
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}
	return err
}

func (c *Client) pollProjectRef(p Project, ref string) {
	var gp *gitlab.Project = p.GitlabProject
	topics := strings.Join(gp.TagList[:], ",")
	var lastPipeline *gitlab.Pipeline

	runCount.WithLabelValues(gp.PathWithNamespace, topics, ref).Add(0)

	b := &backoff.Backoff{
		Min:    time.Duration(cfg.PipelinesPollingIntervalSeconds) * time.Second,
		Max:    time.Duration(cfg.PipelinesMaxPollingIntervalSeconds) * time.Second,
		Factor: 1.4,
		Jitter: false,
	}

	for {
		log.Debugf("Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)

		c.rateLimit()
		pipelines, _, err := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
		if err != nil {
			log.Errorf("ListProjectPipelines: %s", err.Error())
			time.Sleep(b.Duration())
			continue
		}

		if len(pipelines) == 0 {
			log.Debugf("Could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
		} else if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
			if lastPipeline != nil {
				runCount.WithLabelValues(gp.PathWithNamespace, topics, ref).Inc()
			}

			c.rateLimit()
			lastPipeline, _, err = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
			if err != nil {
				log.Errorf("GetPipeline: %s", err.Error())
			}

			if lastPipeline != nil {
				if lastPipeline.Coverage != "" {
					if parsedCoverage, err := strconv.ParseFloat(lastPipeline.Coverage, 64); err == nil {
						coverage.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(parsedCoverage)
					} else {
						log.Warnf("Could not parse coverage string returned from GitLab API: '%s' - '%s'", lastPipeline.Coverage, err.Error())
					}
				}

				lastRunDuration.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(float64(lastPipeline.Duration))
				lastRunID.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(float64(lastPipeline.ID))

				c.outputPipelineStatusMetric(lastPipeline.Status, p.ShouldOutputSparseStatusMetrics(cfg), gp.PathWithNamespace, topics, ref)

				if p.ShouldFetchPipelineJobMetrics(cfg) {
					if err := c.pollPipelineJobs(p, lastPipeline.ID, topics, ref); err != nil {
						log.Errorf("Could not poll jobs for pipeline %d: %s", lastPipeline.ID, err.Error())
					}
				}

			}
		}

		if lastPipeline != nil {
			timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
			b.Reset()
		}

		time.Sleep(b.Duration())
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

func (c *Client) pollProjectRefsFromPipelines(projectID, limit int) ([]string, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: limit,
		},
	}

	refs := []string{}
	c.rateLimit()
	pipelines, _, err := c.Pipelines.ListProjectPipelines(projectID, options)
	if err != nil {
		return refs, err
	}

	for _, p := range pipelines {
		if refExists(refs, p.Ref) {
			continue
		}
		refs = append(refs, p.Ref)
	}

	return refs, nil
}

func (c *Client) rateLimit() {
	now := time.Now()
	throttled := c.RateLimiter.Take()
	if throttled.Sub(now).Milliseconds() > 10 {
		log.Debugf("throttled polling requests for %v", throttled.Sub(now))
	}
}
