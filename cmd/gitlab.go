package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

var statusesList = []string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"}

var has = func(item interface{}, list []interface{}) bool {
	for _, i := range list {
		if cmp.Equal(i, item) {
			return true
		}
	}
	return false
}

func projectExists(p Project, in []Project) bool {
	var ints []interface{}
	for _, pr := range in {
		ints = append(ints, pr)
	}
	return has(p, ints)
}

func refExists(r string, refs []string) bool {
	var ints []interface{}
	for _, ref := range refs {
		ints = append(ints, ref)
	}
	return has(r, ints)
}

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
	RateLimiter ratelimit.Limiter
}

func (c *Client) pollProject(p Project) error {

	log.Debugf("Fetching project : %s", p.Name)
	project, err := c.getProject(p.Name)
	if err != nil {
		return fmt.Errorf("unable to fetch project '%s' from the GitLab API: %v", p.Name, err.Error())
	}
	p.GitlabProject = project

	branchesAndTagRefs, err := c.branchesAndTagsFor(p.GitlabProject.ID, p.Refs)
	if err != nil {
		return fmt.Errorf("error fetching refs for project '%s'", p.Name)
	}
	if len(branchesAndTagRefs) == 0 {
		log.Warnf("No refs found for project '%s'", p.Name)
		return nil
	}
	// read the metrics for refs
	log.Debugf("Polling refs for project : %s", p.Name)
	for _, r := range branchesAndTagRefs {
		log.Infof("Found ref '%s' for project '%s'", r, p.Name)
		if err := c.pollProjectRefOn(project, r, p.ShouldOutputSparseStatusMetrics(cfg), p.ShouldFetchPipelineJobMetrics(cfg)); err != nil {
			log.Errorf("Error getting pipeline data on ref '%s' for project '%s'", r, p.Name)
			continue
		}
	}

	return nil
}

func (c *Client) pollPipelineJobs(gp *gitlab.Project, pipelineID int, topics string, ref string, outputSparseMetricsStatus bool) error {
	var jobs []*gitlab.Job
	var resp *gitlab.Response
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
		// early break if list of jobs is empty
		if jobCount == 0 {
			log.Warnf("No jobs found for pipeline %d", pipelineID)
			break
		}

		//otherwise proceed
		log.Infof("Found %d jobs for pipeline %d", jobCount, pipelineID)
		for _, job := range jobs {
			jobName := job.Name
			stageName := job.Stage
			log.Debugf("Job %s for pipeline %d", jobName, pipelineID)
			lastRunJobDuration.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(job.Duration)

			emitStatusMetric(
				lastRunJobStatus,
				[]string{gp.PathWithNamespace, topics, ref, stageName, jobName},
				statusesList,
				job.Status,
				outputSparseMetricsStatus,
			)

			timeSinceLastJobRun.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())
			jobRunCount.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Inc()

			artifactSize := 0
			for _, artifact := range job.Artifacts {
				artifactSize += artifact.Size
			}

			lastRunJobArtifactSize.WithLabelValues(gp.PathWithNamespace, topics, ref, stageName, jobName).Set(float64(artifactSize))
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}
	return err
}

func (c *Client) pollProjectRefOn(gp *gitlab.Project, ref string, outputSparseStatusMetrics bool, fetchPipelineJobMetrics bool) error {
	pipelines, err := c.pipelinesFor(gp, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", gp.PathWithNamespace, err)
	}
	// fetch pipeline variables, stick them into metrics registry (if requested to do so)
	if cfg.FetchPipelineVariables {
		rx, err := regexp.Compile(cfg.PipelineVariablesFilterRegexp)
		if err != nil {
			log.Errorf("the provided filter regex for pipeline variables is invalid '(%s)': %v", cfg.PipelineVariablesFilterRegexp, err)
			goto process
		}
		for _, pipe := range pipelines {
			if err := emitPipelineVariablesMetric(c, pipelineVariables, gp.PathWithNamespace, ref, gp.ID, pipe.ID, c.Pipelines.GetPipelineVariables, rx); err != nil {
				log.Errorf("%v", err)
			}
		}
	}

process:
	// create the initial matric with topics label, not harmful if it already exists
	topics := strings.Join(gp.TagList[:], ",")
	runCount.WithLabelValues(gp.PathWithNamespace, topics, ref).Add(0)

	if len(pipelines) == 0 {
		return fmt.Errorf("could not find any pipeline run in project %s", gp.PathWithNamespace)
	}
	c.rateLimit()
	lastPipeline, _, err := c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
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

		emitStatusMetric(
			lastRunStatus,
			[]string{gp.PathWithNamespace, topics, ref},
			statusesList,
			lastPipeline.Status,
			outputSparseStatusMetrics,
		)

		if fetchPipelineJobMetrics {
			if err := c.pollPipelineJobs(gp, lastPipeline.ID, topics, ref, outputSparseStatusMetrics); err != nil {
				log.Errorf("Could not poll jobs for pipeline %d: %s", lastPipeline.ID, err.Error())
			}
		}

		timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, topics, ref).Set(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds())
	}

	return nil
}

func (c *Client) orchestratePolling(until <-chan bool, getRefsOnInit <-chan bool) {
	go func() {
		pollWildcardsEvery := time.NewTicker(time.Duration(cfg.ProjectsPollingIntervalSeconds) * time.Second)
		pollRefsEvery := time.NewTicker(time.Duration(cfg.RefsPollingIntervalSeconds) * time.Second)
		stopWorkers := make(chan struct{})
		defer close(stopWorkers)
		// first execution, blocking call before entering orchestration loop to get the wildcards discovered
		c.discoverWildcards()

		for {
			select {
			case <-until:
				log.Info("stopping projects polling...")
				return
			case todo := <-getRefsOnInit:
				log.Debugf("should we poll the more recent refs from the last executed pipelines? %v", todo)
				if todo {
					c.pollPipelinesOnInit()
				}
			case <-pollRefsEvery.C:
				//
				c.pollWithWorkersUntil(stopWorkers)
			case <-pollWildcardsEvery.C:
				// refresh the list of projects from wildcards
				c.discoverWildcards()
			}
		}
	}()
}

func (c *Client) discoverWildcards() {
	log.Infof("%d wildcard(s) configured for polling", len(cfg.Wildcards))
	for _, w := range cfg.Wildcards {
		foundProjects, err := c.listProjects(&w)
		if err != nil {
			log.Errorf("could not list projects for wildcard %v: %v", w, err)
			continue
		}
		for _, p := range foundProjects {
			if !projectExists(p, cfg.Projects) {
				log.Infof("Found new project: %s", p.Name)
				cfg.Projects = append(cfg.Projects, p)
			}
		}
	}
}

func (c *Client) pollPipelinesOnInit() {
	log.Debug("Polling latest pipelines to get data out of them")
	for _, p := range cfg.Projects {
		log.Debugf("On init: reading project %s", p.Name)
		gitlabProject, err := c.getProject(p.Name)
		if err != nil {
			log.Errorf("could not get GitLab project with name %s: %v", p.Name, err)
			continue
		}
		pipelineRefs, err := c.refsFromPipelines(gitlabProject, cfg.OnInitFetchRefsFromPipelinesDepthLimit)
		if err != nil {
			log.Errorf("unable to fetch refs from project pipelines %s : %v", p.Name, err.Error())
			continue
		}
		for _, r := range pipelineRefs {
			if err := c.pollProjectRefOn(gitlabProject, r, false, false); err != nil {
				log.Errorf("%v", err)
			}
		}
	}
}

func (c *Client) pollWithWorkersUntil(stop <-chan struct{}) {
	log.Infof("%d project(s) configured for polling", len(cfg.Projects))
	pollErrors := c.pollProjectsWith(cfg.MaximumProjectsPollingWorkers, c.pollProject, stop, cfg.Projects...)
	for err := range pollErrors {
		if err != nil {
			log.Errorf("%v", err)
		}
	}
}

func (c *Client) pollProjectsWith(numWorkers int, doing func(Project) error, until <-chan struct{}, projects ...Project) <-chan error {
	errorStream := make(chan error)
	projectsToPoll := make(chan Project, len(projects))
	// sync closing the error channel via a waitGroup
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	// spawn maximum_projects_poller_workers to process project polling in parallel
	for w := 0; w < numWorkers; w++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for p := range projectsToPoll {
				select {
				case <-until:
					return
				case errorStream <- doing(p):
				}
			}
		}(&wg)
	}
	// close the error channel when the workers won't write to it anymore
	go func() {
		wg.Wait()
		close(errorStream)
	}()
	// start processing all the projects configured for this run;
	// since the channel is buffered because we already know the length of the projects to process,
	// we can close immediately and the runtime will handle the channel close only when the messages are dispatched
	for _, pr := range projects {
		projectsToPoll <- pr
	}
	close(projectsToPoll)
	return errorStream
}

func (c *Client) pipelinesFor(gp *gitlab.Project, options *gitlab.ListProjectPipelinesOptions) ([]*gitlab.PipelineInfo, error) {
	log.Debugf("Reading pipelines for project %v (ID %v)", gp.PathWithNamespace, gp.ID)
	c.rateLimit()
	pipelines, _, err := c.Pipelines.ListProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project %s: %v", gp.PathWithNamespace, err)
	}
	return pipelines, nil
}

func (c *Client) refsFromPipelines(gp *gitlab.Project, limit int) ([]string, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: limit,
		},
	}
	pipelines, err := c.pipelinesFor(gp, options)
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, p := range pipelines {
		if !refExists(p.Ref, refs) {
			refs = append(refs, p.Ref)
		}
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

func (c *Client) branchesAndTagsFor(projectID int, refsRegexp string) ([]string, error) {
	if len(refsRegexp) == 0 {
		if len(cfg.DefaultRefsRegexp) > 0 {
			refsRegexp = cfg.DefaultRefsRegexp
		} else {
			refsRegexp = "^master$"
		}
	}
	var refs []string
	re := regexp.MustCompile(refsRegexp)
	branches, err := c.branchNamesFor(projectID)
	if err != nil {
		return nil, err
	}
	tags, err := c.tagNamesFor(projectID)
	if err != nil {
		return nil, err
	}
	for _, ref := range append(branches, tags...) {
		if re.MatchString(*ref) {
			refs = append(refs, *ref)
		}
	}
	return refs, nil
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
