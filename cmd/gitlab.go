package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
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

	var projects []Project
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

func (c *Client) branchesAndTagsFor(projectID int, refsRegexp string) (refs []string, err error) {
	if len(refsRegexp) == 0 {
		if len(cfg.DefaultRefsRegexp) > 0 {
			refsRegexp = cfg.DefaultRefsRegexp
		} else {
			refsRegexp = "^master$"
		}
	}

	re := regexp.MustCompile(refsRegexp)

	branches, err := c.branchNamesFor(projectID)
	if err != nil {
		return
	}

	for _, branch := range branches {
		if re.MatchString(*branch) {
			refs = append(refs, *branch)
		}
	}

	tags, err := c.tagNamesFor(projectID)
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

func (c *Client) branchNamesFor(projectID int) ([]*string, error) {
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

func (c *Client) tagNamesFor(projectID int) ([]*string, error) {
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

func (c *Client) pollProject(p Project) error {
	var polledRefs []string
	var err error

	log.Debugf("Fetching project : %s", p.Name)
	p.GitlabProject, err = c.getProject(p.Name)
	if err != nil {
		return fmt.Errorf("unable to fetch project '%s' from the GitLab API : %v", p.Name, err.Error())
	}

	if cfg.OnInitFetchRefsFromPipelines {
		log.Debugf("Polling project refs %s from most recent %d pipelines (init only)", p.Name, cfg.OnInitFetchRefsFromPipelinesDepthLimit)
		refs, err := c.pollProjectRefsFromPipelines(p.GitlabProject.ID, cfg.OnInitFetchRefsFromPipelinesDepthLimit)
		if err != nil {
			return fmt.Errorf("unable to fetch refs from project pipelines %s : %v", p.Name, err.Error())
		}

		log.Debugf("Found %d refs from %s pipelines", len(refs), p.Name)
		for _, r := range refs {
			log.Infof("Found ref '%s' for project '%s'", r, p.Name)
			if err := c.pollProjectRef(p, r); err != nil {
				log.Errorf("Error getting pipeline data on ref '%s' for project '%s'", r, p.Name)
				continue
			}
			polledRefs = append(polledRefs, r)
		}
	}
	log.Debugf("Polling refs for project : %s", p.Name)

	refs, err := c.branchesAndTagsFor(p.GitlabProject.ID, p.Refs)
	if err != nil {
		return fmt.Errorf("error fetching refs for project '%s'", p.Name)
	}

	if len(refs) > 0 {
		for _, r := range refs {
			if !refExists(polledRefs, r) {
				log.Infof("Found ref '%s' for project '%s'", r, p.Name)
				if err := c.pollProjectRef(p, r); err != nil {
					log.Errorf("Error getting pipeline data on ref '%s' for project '%s'", r, p.Name)
					continue
				}
				polledRefs = append(polledRefs, r)
			}
		}
	}

	log.Warnf("No refs found for project '%s'", p.Name)
	return nil
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
	var gp = p.GitlabProject
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
				lastRunJobDuration.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(job.Duration)

				c.outputJobStatusMetric(job.Status, p.ShouldOutputSparseStatusMetrics(cfg), gp.PathWithNamespace, topics, ref, stageName, jobName)

				timeSinceLastJobRun.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())

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

func (c *Client) pollProjectRef(p Project, ref string) error {
	var gp = p.GitlabProject
	var lastPipeline *gitlab.Pipeline
	topics := strings.Join(gp.TagList[:], ",")

	runCount.WithLabelValues(gp.PathWithNamespace, topics, ref).Add(0)
	log.Debugf("Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)

	c.rateLimit()
	pipelines, _, err := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
	if err != nil {
		return fmt.Errorf("error listing project pipelines for ref %s: %v", ref, err)
	}

	if len(pipelines) == 0 {
		return fmt.Errorf("could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
	}

	c.rateLimit()
	lastPipeline, _, err = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
	if err != nil {
		return fmt.Errorf("could not read content of last pipeline %s:%s", gp.PathWithNamespace, ref)
	}
	if lastPipeline != nil {
		runCount.WithLabelValues(gp.PathWithNamespace, topics, ref).Inc()

		if lastPipeline.Coverage != "" {
			parsedCoverage, err := strconv.ParseFloat(lastPipeline.Coverage, 64)
			if err != nil {
				log.Warnf("Could not parse coverage string returned from GitLab API '%s' into Float64: %v", lastPipeline.Coverage, err)
			}
			coverage.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(parsedCoverage)
		}

		lastRunDuration.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(float64(lastPipeline.Duration))
		lastRunID.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(float64(lastPipeline.ID))

		c.outputPipelineStatusMetric(lastPipeline.Status, p.ShouldOutputSparseStatusMetrics(cfg), gp.PathWithNamespace, topics, ref)

		if p.ShouldFetchPipelineJobMetrics(cfg) {
			if err := c.pollPipelineJobs(p, lastPipeline.ID, topics, ref); err != nil {
				log.Errorf("Could not poll jobs for pipeline %d: %s", lastPipeline.ID, err.Error())
			}
		}

		timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds())
	}

	return nil
}

func (c *Client) pollProjects(until <-chan bool) {
	go func() {
		pollWildcardsEvery := time.NewTicker(time.Duration(cfg.ProjectsPollingIntervalSeconds) * time.Second)
		pollRefsEvery := time.NewTicker(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
		stopWorkers := make(chan struct{})
		defer close(stopWorkers)
		// first execution
		c.discoverWildcards()
		c.pollWithWorkersUntil(stopWorkers)
		for {
			select {
			case <-until:
				log.Info("stopping projects polling...")
				return
			case <-pollRefsEvery.C:
				c.pollWithWorkersUntil(stopWorkers)
			case <-pollWildcardsEvery.C:
				c.discoverWildcards()
			}
		}
	}()
}

func (c *Client) discoverWildcards() {
	log.Infof("%d wildcard(s) configured for polling", len(cfg.Wildcards))
	if err := c.findProjectsFromWildcards(); err != nil {
		log.Errorf("%v", err)
	}
}

func (c *Client) pollWithWorkersUntil(stop <-chan struct{}) {
	log.Infof("%d project(s) configured for polling", len(cfg.Projects))
	pollErrors := c.pollProjectsWith(cfg.MaximumProjectsPollingWorkers, stop, cfg.Projects...)
	for _, err := range pollErrors {
		log.Errorf("%v", err)
	}
}

func (c *Client) findProjectsFromWildcards() error {
	for _, w := range cfg.Wildcards {
		foundProjects, err := c.listProjects(&w)
		if err != nil {
			return fmt.Errorf("could not list projects for wildcard %v: %v", w, err)
		}
		for _, p := range foundProjects {
			if !projectExists(p) {
				log.Infof("Found new project: %s", p.Name)
				cfg.Projects = append(cfg.Projects, p)
			}
		}
	}
	return nil
}

func (c *Client) pollProjectsWith(numWorkers int, until <-chan struct{}, projects ...Project) []error {
	var errs []error
	errorStream := make(chan error)
	defer close(errorStream)
	projectsToPoll := make(chan Project, len(projects))
	// spawn maximum_projects_poller_workers to process project polling in parallel
	for w := 0; w < numWorkers; w++ {
		go func() {
			for {
				select {
				case <-until:
					return
				case p := <-projectsToPoll:
					if e := c.pollProject(p); e != nil {
						errorStream <- e
					}
				}
			}
		}()
	}
	// process errors coming from pollProject
	go func() {
		for ex := range errorStream {
			errs = append(errs, ex)
		}
	}()
	// start processing all the projects configured for this run;
	// since the channel is buffered because we already know the length of the projects to process,
	// we can close immediately and the runtime will handle the channel close only when the messages are dispatched
	for _, pr := range projects {
		projectsToPoll <- pr
	}
	close(projectsToPoll)
	return errs
}

func (c *Client) pollProjectRefsFromPipelines(projectID, limit int) ([]string, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: limit,
		},
	}

	var refs []string
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
