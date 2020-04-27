package exporter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/schemas"
)

var statusesList = []string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"}

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
	Config      *schemas.Config
	RateLimiter ratelimit.Limiter
}

type projectDetails struct {
	*schemas.Project

	fullName, topics, ref string
	pID                   int
}

func refExists(r string, refs []string) bool {
	for _, ref := range refs {
		if ref == r {
			return true
		}
	}
	return false
}

func detailFrom(p *schemas.Project, gp *gitlab.Project, ref string) *projectDetails {
	topics := strings.Join(gp.TagList, ",")
	return &projectDetails{
		p, gp.PathWithNamespace, topics, ref, gp.ID,
	}
}

func (d *projectDetails) stdLabelValues() []string {
	return []string{d.fullName, d.topics, d.ref}
}

func augmentLabelValues(d *projectDetails, lbv ...string) []string {
	std := d.stdLabelValues()
	for _, l := range lbv {
		std = append(std, l)
	}
	return std
}

func (c *Client) pollProject(p schemas.Project) error {
	log.Debugf("Fetching project : %s", p.Name)
	project, err := c.getProject(p.Name)
	if err != nil {
		return fmt.Errorf("unable to fetch project '%s' from the GitLab API: %v", p.Name, err.Error())
	}

	branchesAndTagRefs, err := c.branchesAndTagsFor(project.ID, p.RefsRegexp(c.Config))
	if err != nil {
		return fmt.Errorf("error fetching refs for project '%s'", p.Name)
	}
	if len(branchesAndTagRefs) == 0 {
		log.Warnf("No refs found for project '%s'", p.Name)
		return nil
	}
	// read the metrics for refs
	log.Debugf("Polling refs for project : %s", p.Name)
	for _, ref := range branchesAndTagRefs {
		pd := detailFrom(&p, project, ref)
		log.Infof("Found ref '%s' for project '%s'", ref, pd.fullName)
		if err := c.pollProjectRefOn(pd); err != nil {
			log.Errorf("Error getting pipeline data on ref '%s' for project '%s': %v", ref, p.Name, err)
			continue
		}
	}

	return nil
}

func (c *Client) pollPipelineJobs(pd *projectDetails, pipelineID int) error {
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
		jobs, resp, err = c.Jobs.ListPipelineJobs(pd.pID, pipelineID, options)
		if err != nil {
			return err
		}

		//otherwise proceed
		log.Infof("Found %d jobs for pipeline %d", len(jobs), pipelineID)
		for _, job := range jobs {
			var values = augmentLabelValues(pd, job.Stage, job.Name)

			log.Debugf("Job %s for pipeline %d", job.Name, pipelineID)
			lastRunJobDuration.WithLabelValues(values...).Set(job.Duration)

			emitStatusMetric(
				lastRunJobStatus,
				values,
				statusesList,
				job.Status,
				pd.OutputSparseStatusMetrics(c.Config),
			)

			timeSinceLastJobRun.WithLabelValues(values...).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())
			jobRunCount.WithLabelValues(values...).Inc()

			artifactSize := 0
			for _, artifact := range job.Artifacts {
				artifactSize += artifact.Size
			}

			lastRunJobArtifactSize.WithLabelValues(values...).Set(float64(artifactSize))
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}
	return err
}

func (c *Client) pollProjectRefOn(pd *projectDetails) error {
	pipelines, err := c.pipelinesFor(pd, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(pd.ref)})
	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", pd.fullName, err)
	}

	// fetch pipeline variables, stick them into metrics registry (if requested to do so)
	if pd.FetchPipelineVariables(c.Config) {
		rx, err := regexp.Compile(pd.PipelineVariablesRegexp(c.Config))
		if err != nil {
			log.Errorf("The provided filter regex for pipeline variables is invalid '(%s)': %v", pd.PipelineVariablesRegexp(c.Config), err)
			goto process
		}
		for _, pipe := range pipelines {
			if err := emitPipelineVariablesMetric(c, pipelineVariables, pd, pipe.ID, c.Pipelines.GetPipelineVariables, rx); err != nil {
				log.Errorf("%v", err)
			}
		}
	}

process:
	if len(pipelines) == 0 {
		return fmt.Errorf("could not find any pipeline run in project %s", pd.fullName)
	}
	c.rateLimit()
	lastPipeline, _, err := c.Pipelines.GetPipeline(pd.pID, pipelines[0].ID)
	if err != nil {
		return fmt.Errorf("could not read content of last pipeline %s:%s", pd.fullName, pd.ref)
	}
	if lastPipeline != nil {
		var labelValues = pd.stdLabelValues()
		runCount.WithLabelValues(labelValues...).Inc()

		if lastPipeline.Coverage != "" {
			parsedCoverage, err := strconv.ParseFloat(lastPipeline.Coverage, 64)
			if err != nil {
				log.Warnf("Could not parse coverage string returned from GitLab API '%s' into Float64: %v", lastPipeline.Coverage, err)
			}
			coverage.WithLabelValues(labelValues...).Set(parsedCoverage)
		}

		lastRunDuration.WithLabelValues(labelValues...).Set(float64(lastPipeline.Duration))
		lastRunID.WithLabelValues(labelValues...).Set(float64(lastPipeline.ID))

		emitStatusMetric(
			lastRunStatus,
			labelValues,
			statusesList,
			lastPipeline.Status,
			pd.OutputSparseStatusMetrics(c.Config),
		)

		if pd.FetchPipelineJobMetrics(c.Config) {
			if err := c.pollPipelineJobs(pd, lastPipeline.ID); err != nil {
				log.Errorf("Could not poll jobs for pipeline %d: %s", lastPipeline.ID, err.Error())
			}
		}

		timeSinceLastRun.WithLabelValues(labelValues...).Set(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds())
	}

	return nil
}

// OrchestratePolling ...
func (c *Client) OrchestratePolling(until <-chan bool, getRefsOnInit <-chan bool) {
	go func() {
		pollWildcardsEvery := time.NewTicker(time.Duration(c.Config.ProjectsPollingIntervalSeconds) * time.Second)
		pollRefsEvery := time.NewTicker(time.Duration(c.Config.RefsPollingIntervalSeconds) * time.Second)
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
	log.Infof("%d wildcard(s) configured for polling", len(c.Config.Wildcards))
	for _, w := range c.Config.Wildcards {
		foundProjects, err := c.listProjects(&w)
		if err != nil {
			log.Errorf("could not list projects for wildcard %v: %v", w, err)
			continue
		}
		for _, p := range foundProjects {
			if !c.Config.ProjectExists(p) {
				log.Infof("Found new project: %s", p.Name)
				c.Config.Projects = append(c.Config.Projects, p)
			}
		}
	}
}

func (c *Client) pollPipelinesOnInit() {
	log.Debug("Polling latest pipelines to get data out of them")
	for _, p := range c.Config.Projects {
		log.Debugf("On init: reading project %s", p.Name)
		gitlabProject, err := c.getProject(p.Name)
		if err != nil {
			log.Errorf("could not get GitLab project with name %s: %v", p.Name, err)
			continue
		}
		pipelineRefs, err := c.refsFromPipelines(detailFrom(&p, gitlabProject, ""))
		if err != nil {
			log.Errorf("unable to fetch refs from project pipelines %s : %v", p.Name, err.Error())
			continue
		}
		for _, ref := range pipelineRefs {
			if err := c.pollProjectRefOn(detailFrom(&p, gitlabProject, ref)); err != nil {
				log.Errorf("%v", err)
			}
		}
	}
}

func (c *Client) pollWithWorkersUntil(stop <-chan struct{}) {
	log.Infof("%d project(s) configured for polling", len(c.Config.Projects))
	pollErrors := pollProjectsWith(c.Config.MaximumProjectsPollingWorkers, c.pollProject, stop, c.Config.Projects...)
	for err := range pollErrors {
		if err != nil {
			log.Errorf("%v", err)
		}
	}
}

func pollProjectsWith(numWorkers int, doing func(schemas.Project) error, until <-chan struct{}, projects ...schemas.Project) <-chan error {
	errorStream := make(chan error)
	projectsToPoll := make(chan schemas.Project, len(projects))
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

func (c *Client) pipelinesFor(pd *projectDetails, options *gitlab.ListProjectPipelinesOptions) ([]*gitlab.PipelineInfo, error) {
	log.Debugf("Reading pipelines for project %v (ID %v)", pd.fullName, pd.pID)
	c.rateLimit()

	pipelines, _, err := c.Pipelines.ListProjectPipelines(pd.pID, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project %s: %v", pd.fullName, err)
	}
	return pipelines, nil
}

func (c *Client) refsFromPipelines(pd *projectDetails) ([]string, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: c.Config.OnInitFetchRefsFromPipelinesDepthLimit,
		},
	}
	pipelines, err := c.pipelinesFor(pd, options)
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

func (c *Client) listProjects(w *schemas.Wildcard) ([]schemas.Project, error) {
	log.Debugf("Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

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

func (c *Client) branchesAndTagsFor(projectID int, refsRegexp string) ([]string, error) {
	var refs []string
	re, err := regexp.Compile(refsRegexp)
	if err != nil {
		return nil, err
	}

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
