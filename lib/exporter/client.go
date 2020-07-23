package exporter

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openlyinc/pointy"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/lib/schemas"
)

var statusesList = [...]string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"}

// ProjectRefKind is used to determine the kind of the ref
type ProjectRefKind string

const (
	// ProjectRefKindBranch refers to a branch
	ProjectRefKindBranch ProjectRefKind = "branch"

	// ProjectRefKindTag refers to a tag
	ProjectRefKindTag ProjectRefKind = "tag"

	// ProjectRefKindMergeRequest refers to a tag
	ProjectRefKindMergeRequest ProjectRefKind = "merge-request"

	mergeRequestRefRegexp = `^refs/merge-requests`
)

// ProjectRef is what we will use a metrics entity on which we will
// perform regular polling operations
type ProjectRef struct {
	*schemas.Project

	Kind                          ProjectRefKind
	ID                            int
	PathWithNamespace             string
	Topics                        string
	Ref                           string
	MostRecentPipeline            *gitlab.Pipeline
	MostRecentPipelineVariables   string
	PreviouslyEmittedPipelineJobs map[string]int
}

// ProjectsRefs allows us to keep track of all the ProjectRef
// we have configured/discovered
type ProjectsRefs map[int]map[string]*ProjectRef

// Count returns the amount of projects refs in the map
func (prs ProjectsRefs) Count() (count int) {
	for _, projectRefs := range prs {
		count += len(projectRefs)
	}
	return
}

// NewProjectRef is an helper which returns a new ProjectRef pointer
func NewProjectRef(project *schemas.Project, gp *gitlab.Project, ref string, kind ProjectRefKind) *ProjectRef {
	return &ProjectRef{
		Project:           project,
		Kind:              kind,
		ID:                gp.ID,
		PathWithNamespace: gp.PathWithNamespace,
		Topics:            strings.Join(gp.TagList, ","),
		Ref:               ref,
	}
}

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
	Config       *schemas.Config
	RateLimiter  ratelimit.Limiter
	ProjectsRefs ProjectsRefs
	Mutex        sync.Mutex
}

func (pr *ProjectRef) defaultLabelsValues() []string {
	return []string{pr.PathWithNamespace, pr.Topics, pr.Ref, string(pr.Kind), pr.MostRecentPipelineVariables}
}

// runWithContext wraps polling function in order to be able to gracefully exit during potentially
// long running operations.
func runWithContext(ctx context.Context, f func(), functionName string) {
	exit := make(chan bool)
	go func(exit chan<- bool) {
		f()
		exit <- true
	}(exit)

	select {
	case <-ctx.Done():
		log.Infof("stopped %s", functionName)
	case <-exit:
		log.Debugf("finished %s", functionName)
	}
}

// OrchestratePolling ...
func (c *Client) OrchestratePolling(ctx context.Context) {
	go func(ctx context.Context) {
		discoverWildcardsProjectsEvery := time.NewTicker(time.Duration(c.Config.WildcardsProjectsDiscoverIntervalSeconds) * time.Second)
		discoverProjectsRefsEvery := time.NewTicker(time.Duration(c.Config.ProjectsRefsDiscoverIntervalSeconds) * time.Second)
		pollProjectsRefsEvery := time.NewTicker(time.Duration(c.Config.ProjectsRefsPollingIntervalSeconds) * time.Second)

		// first execution, blocking call to initiate everything before entering the orchestration loop
		runWithContext(ctx, c.discoverProjectsRefsFromPipelinesOnInit, "discoverProjectsRefsFromPipelinesOnInit")
		runWithContext(ctx, c.discoverProjectsFromWildcards, "discoverProjectsFromWildcards")
		runWithContext(ctx, c.discoverProjectsRefs, "discoverProjectsRefs")
		c.pollProjectsRefs(ctx, c.pollProjectRefMostRecentPipeline)

		// Then, waiting for the tickers to kick in
		for {
			select {
			case <-ctx.Done():
				log.Info("stopped polling orchestration")
				return
			case <-discoverWildcardsProjectsEvery.C:
				runWithContext(ctx, c.discoverProjectsFromWildcards, "discoverProjectsFromWildcards")
			case <-discoverProjectsRefsEvery.C:
				runWithContext(ctx, c.discoverProjectsRefs, "discoverProjectsRefs")
			case <-pollProjectsRefsEvery.C:
				c.pollProjectsRefs(ctx, c.pollProjectRefMostRecentPipeline)
			}
		}
	}(ctx)
}

func (c *Client) pollPipelineJobs(pr *ProjectRef) error {
	var jobs []*gitlab.Job
	var resp *gitlab.Response
	var err error

	// Initialize the variable
	if pr.PreviouslyEmittedPipelineJobs == nil {
		pr.PreviouslyEmittedPipelineJobs = map[string]int{}
	}

	options := &gitlab.ListJobsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		c.rateLimit()
		jobs, resp, err = c.Jobs.ListPipelineJobs(pr.ID, pr.MostRecentPipeline.ID, options)
		if err != nil {
			return err
		}

		// otherwise proceed
		log.WithFields(
			log.Fields{
				"project-id":  pr.ID,
				"pipeline-id": pr.MostRecentPipeline.ID,
				"jobs-count":  len(jobs),
			},
		).Info("found pipeline jobs")

		for _, job := range jobs {
			jobValues := append(pr.defaultLabelsValues(), job.Stage, job.Name)

			// In case a job gets restarted, it will have an ID greated than the previous one(s)
			// jobs in new pipelines should get greated IDs too
			if previousJobID, ok := pr.PreviouslyEmittedPipelineJobs[job.Name]; ok {
				if previousJobID == job.ID {
					timeSinceLastJobRun.WithLabelValues(jobValues...).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())
				}

				if previousJobID >= job.ID {
					continue
				}
			}

			log.WithFields(
				log.Fields{
					"project-id":  pr.ID,
					"pipeline-id": pr.MostRecentPipeline.ID,
					"job-name":    job.Name,
					"job-id":      job.ID,
				},
			).Debug("processing pipeline job metrics")

			pr.PreviouslyEmittedPipelineJobs[job.Name] = job.ID

			timeSinceLastJobRun.WithLabelValues(jobValues...).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())
			lastRunJobDuration.WithLabelValues(jobValues...).Set(job.Duration)

			emitStatusMetric(
				lastRunJobStatus,
				jobValues,
				statusesList[:],
				job.Status,
				pr.OutputSparseStatusMetrics(c.Config),
			)

			timeSinceLastJobRun.WithLabelValues(jobValues...).Set(time.Since(*job.CreatedAt).Round(time.Second).Seconds())
			jobRunCount.WithLabelValues(jobValues...).Inc()

			artifactSize := 0
			for _, artifact := range job.Artifacts {
				artifactSize += artifact.Size
			}

			lastRunJobArtifactSize.WithLabelValues(jobValues...).Set(float64(artifactSize))
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}
	return err
}

func (c *Client) pollProjectRefMostRecentPipeline(pr *ProjectRef) error {
	// TODO: Figure out if we want to have a similar approach for ProjectRefKindTag with
	// an additional configuration parameeter perhaps
	if pr.Kind == ProjectRefKindMergeRequest && pr.MostRecentPipeline != nil {
		switch pr.MostRecentPipeline.Status {
		case "success", "failed", "canceled", "skipped":
			// The pipeline will not evolve, lets not bother querying the API
			log.WithFields(
				log.Fields{
					"project-path-with-namespace": pr.PathWithNamespace,
					"project-id":                  pr.ID,
					"project-ref":                 pr.Ref,
					"project-ref-kind":            pr.Kind,
					"pipeline-id":                 pr.MostRecentPipeline.ID,
				},
			).Debug("skipping finished merge-request pipeline")
			return nil
		}
	}

	pipelines, err := c.getProjectPipelines(pr.ID, &gitlab.ListProjectPipelinesOptions{
		// We only need the most recent pipeline
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
		Ref: gitlab.String(pr.Ref),
	})

	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", pr.PathWithNamespace, err)
	}

	if len(pipelines) == 0 {
		return fmt.Errorf("could not find any pipeline for project %s with ref %s", pr.PathWithNamespace, pr.Ref)
	}

	c.rateLimit()
	pipeline, _, err := c.Pipelines.GetPipeline(pr.ID, pipelines[0].ID)
	if err != nil || pipeline == nil {
		return fmt.Errorf("could not read content of last pipeline %s:%s", pr.PathWithNamespace, pr.Ref)
	}

	defaultLabelValues := pr.defaultLabelsValues()
	if pr.MostRecentPipeline == nil || !reflect.DeepEqual(pipeline, pr.MostRecentPipeline) {
		pr.MostRecentPipeline = pipeline

		// fetch pipeline variables
		if pr.FetchPipelineVariables(c.Config) {
			pr.MostRecentPipelineVariables, err = c.getPipelineVariablesAsConcatenatedString(pr)
			if err != nil {
				return err
			}
		} else {
			// Ensure we flush the value if there was some variables defined on the previous pipeline
			pr.MostRecentPipelineVariables = ""
		}

		if pipeline.Status == "running" {
			runCount.WithLabelValues(defaultLabelValues...).Inc()
		}

		if pipeline.Coverage != "" {
			parsedCoverage, err := strconv.ParseFloat(pipeline.Coverage, 64)
			if err != nil {
				log.Warnf("Could not parse coverage string returned from GitLab API '%s' into Float64: %v", pipeline.Coverage, err)
			}
			coverage.WithLabelValues(defaultLabelValues...).Set(parsedCoverage)
		}

		lastRunDuration.WithLabelValues(defaultLabelValues...).Set(float64(pipeline.Duration))
		lastRunID.WithLabelValues(defaultLabelValues...).Set(float64(pipeline.ID))

		emitStatusMetric(
			lastRunStatus,
			defaultLabelValues,
			statusesList[:],
			pipeline.Status,
			pr.OutputSparseStatusMetrics(c.Config),
		)
	}

	if pr.FetchPipelineJobMetrics(c.Config) {
		if err := c.pollPipelineJobs(pr); err != nil {
			log.WithFields(
				log.Fields{
					"project-path-with-namespace": pr.PathWithNamespace,
					"project-id":                  pr.ID,
					"project-ref":                 pr.Ref,
					"project-ref-kind":            pr.Kind,
					"pipeline-id":                 pipeline.ID,
					"error":                       err.Error(),
				},
			).Error("polling pipeline jobs metrics")
		}
	}

	timeSinceLastRun.WithLabelValues(defaultLabelValues...).Set(time.Since(*pipeline.CreatedAt).Round(time.Second).Seconds())

	return nil
}

func (c *Client) discoverProjectsFromWildcards() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	log.WithFields(
		log.Fields{
			"total": len(c.Config.Wildcards),
		},
	).Info("discover wildcards")

	for _, w := range c.Config.Wildcards {
		foundProjects, err := c.listProjects(&w)
		if err != nil {
			log.WithFields(
				log.Fields{
					"wildcard-search":                  w.Search,
					"wildcard-owner-kind":              w.Owner.Kind,
					"wildcard-owner-name":              w.Owner.Name,
					"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
					"wildcard-archived":                w.Archived,
					"error":                            err.Error(),
				},
			).Errorf("listing wildcard projects")
			continue
		}
		for _, p := range foundProjects {
			if !c.Config.ProjectExists(p) {
				log.WithFields(
					log.Fields{
						"wildcard-search":                  w.Search,
						"wildcard-owner-kind":              w.Owner.Kind,
						"wildcard-owner-name":              w.Owner.Name,
						"wildcard-owner-include-subgroups": w.Owner.IncludeSubgroups,
						"wildcard-archived":                w.Archived,
						"project-name":                     p.Name,
					},
				).Infof("discovered new project")
				c.Config.Projects = append(c.Config.Projects, p)
			}
		}
	}
}

func (c *Client) discoverProjectsRefsFromPipelinesOnInit() {
	if !c.Config.OnInitFetchRefsFromPipelines {
		log.WithFields(
			log.Fields{
				"init-operation": true,
			},
		).Debug("not configured to fetch refs from most recent pipelines")
		return
	}

	log.WithFields(
		log.Fields{
			"init-operation": true,
		},
	).Debug("polling most recent project pipelines")

	c.ProjectsRefs = ProjectsRefs{}
	for _, p := range c.Config.Projects {
		log.WithFields(
			log.Fields{
				"init-operation": true,
				"project-name":   p.Name,
			},
		).Debug("fetching project")

		gp, err := c.getProject(p.Name)
		if err != nil {
			log.WithFields(
				log.Fields{
					"init-operation": true,
					"project-name":   p.Name,
					"error":          err.Error(),
				},
			).Errorf("getting GitLab project by name")
			continue
		}

		c.ProjectsRefs[gp.ID], err = c.getProjectsRefsFromPipelines(&p, gp)
		if err != nil {
			log.WithFields(
				log.Fields{
					"init-operation": true,
					"project-name":   p.Name,
				},
			).Errorf("unable to fetch refs from project pipelines: %s", err.Error())
		}
	}
}

func (c *Client) pollProjectsRefs(ctx context.Context, pollFunction func(*ProjectRef) error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	log.WithFields(
		log.Fields{
			"total": c.ProjectsRefs.Count(),
		},
	).Info("polling metrics from projects refs")

	projectsRefsToPoll := make(chan *ProjectRef)

	// sync closing the error channel via a waitGroup
	var wg sync.WaitGroup
	wg.Add(c.Config.PollingWorkers)
	defer wg.Wait()

	// spawn workers to process project polling in parallel
	for workerID := 0; workerID < c.Config.PollingWorkers; workerID++ {
		log.WithFields(
			log.Fields{
				"polling-worker-id": workerID,
			},
		).Debug("polling worker started")

		go func(ctx context.Context, workerID int, wg *sync.WaitGroup, projectsRefsToPoll <-chan *ProjectRef) {
			defer func() {
				wg.Done()
				log.WithFields(
					log.Fields{
						"polling-worker-id": workerID,
					},
				).Debug("polling worker stopped")
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case pr, open := <-projectsRefsToPoll:
					if !open {
						return
					}
					if err := pollFunction(pr); err != nil {
						log.WithFields(
							log.Fields{
								"project-path-with-namespace": pr.PathWithNamespace,
								"project-id":                  pr.ID,
								"project-ref":                 pr.Ref,
								"project-ref-kind":            pr.Kind,
								"error":                       err.Error(),
							},
						).Error("whilst metrics from polling projects refs")
					}
				}
			}
		}(ctx, workerID, &wg, projectsRefsToPoll)
	}

	// start processing all the projects configured for this run;
	// since the channel is buffered because we already know the length of the projects to process,
	// we can close immediately and the runtime will handle the channel close only when the messages are dispatched
	for _, refs := range c.ProjectsRefs {
		for _, pr := range refs {
			projectsRefsToPoll <- pr
		}
	}

	close(projectsRefsToPoll)
}

func (c *Client) getProjectPipelines(projectID int, options *gitlab.ListProjectPipelinesOptions) ([]*gitlab.PipelineInfo, error) {
	fields := log.Fields{
		"project-id": projectID,
	}

	if options.Ref != nil {
		fields["project-ref"] = *options.Ref
	}

	log.WithFields(fields).Debug("listing project pipelines")

	c.rateLimit()

	pipelines, _, err := c.Pipelines.ListProjectPipelines(projectID, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project ID %d: %s", projectID, err.Error())
	}
	return pipelines, nil
}

func (c *Client) getProjectsRefsFromPipelines(p *schemas.Project, gp *gitlab.Project) (map[string]*ProjectRef, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
			// TODO: Get a proper loop to split this query up
			PerPage: c.Config.OnInitFetchRefsFromPipelinesDepthLimit,
		},
		Scope: pointy.String("branches"),
	}

	branchPipelines, err := c.getProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, err
	}

	options.Scope = pointy.String("tags")
	tagsPipelines, err := c.getProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(p.RefsRegexp(c.Config))
	if err != nil {
		return nil, err
	}

	projectRefs := map[string]*ProjectRef{}
	for kind, pipelines := range map[ProjectRefKind][]*gitlab.PipelineInfo{
		ProjectRefKindBranch: branchPipelines,
		ProjectRefKindTag:    tagsPipelines,
	} {
		for _, pipeline := range pipelines {
			if re.MatchString(pipeline.Ref) {
				if _, ok := projectRefs[pipeline.Ref]; !ok {
					c.rateLimit()

					log.WithFields(
						log.Fields{
							"project-id":                  gp.ID,
							"project-path-with-namespace": gp.PathWithNamespace,
							"project-ref":                 pipeline.Ref,
							"project-ref-kind":            kind,
						},
					).Info("found project ref")
					projectRefs[pipeline.Ref] = NewProjectRef(p, gp, pipeline.Ref, kind)
				}
			}
		}
	}

	return projectRefs, nil
}

func (c *Client) rateLimit() {
	now := time.Now()
	throttled := c.RateLimiter.Take()
	if throttled.Sub(now).Milliseconds() > 10 {
		log.WithFields(
			log.Fields{
				"for": throttled.Sub(now),
			},
		).Debug("throttled GitLab requests")
	}
}

func (c *Client) getProject(name string) (*gitlab.Project, error) {
	c.rateLimit()
	p, _, err := c.Projects.GetProject(name, &gitlab.GetProjectOptions{})
	return p, err
}

func (c *Client) listProjects(w *schemas.Wildcard) ([]schemas.Project, error) {
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

func (c *Client) discoverProjectsRefs() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.ProjectsRefs == nil {
		c.ProjectsRefs = ProjectsRefs{}
	}

	for _, p := range c.Config.Projects {
		gp, err := c.getProject(p.Name)
		if err != nil {
			log.WithFields(
				log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				},
			).Errorf("getting project by name")
			continue
		}

		if _, ok := c.ProjectsRefs[gp.ID]; !ok {
			c.ProjectsRefs[gp.ID] = map[string]*ProjectRef{}
		}

		refs, err := c.getProjectRefs(
			gp.ID,
			p.RefsRegexp(c.Config),
			p.FetchMergeRequestsPipelinesRefs(c.Config),
			p.FetchMergeRequestsPipelinesRefsLimit(c.Config),
		)

		if err != nil {
			log.WithFields(
				log.Fields{
					"project-id":                  gp.ID,
					"project-path-with-namespace": gp.PathWithNamespace,
					"error":                       err.Error(),
				},
			).Error("getting project refs")
		}

		for ref, kind := range refs {
			if _, ok := c.ProjectsRefs[gp.ID][ref]; !ok {
				log.WithFields(
					log.Fields{
						"project-id":                  gp.ID,
						"project-path-with-namespace": gp.PathWithNamespace,
						"project-ref":                 ref,
						"project-ref-kind":            kind,
					},
				).Info("discovered new project ref")
				c.ProjectsRefs[gp.ID][ref] = NewProjectRef(&p, gp, ref, kind)
			}
		}
	}
}

func (c *Client) getProjectRefs(
	projectID int,
	refsRegexp string,
	fetchMergeRequestsPipelinesRefs bool,
	fetchMergeRequestsPipelinesRefsInitLimit int) (map[string]ProjectRefKind, error) {

	branches, err := c.getProjectBranches(projectID, refsRegexp)
	if err != nil {
		return nil, err
	}

	tags, err := c.getProjectTags(projectID, refsRegexp)
	if err != nil {
		return nil, err
	}

	mergeRequests := []string{}
	if fetchMergeRequestsPipelinesRefs {
		mergeRequests, err = c.getProjectMergeRequestsPipelines(projectID, fetchMergeRequestsPipelinesRefsInitLimit)
		if err != nil {
			return nil, err
		}
	}

	foundRefs := map[string]ProjectRefKind{}
	for kind, refs := range map[ProjectRefKind][]string{
		ProjectRefKindBranch:       branches,
		ProjectRefKindTag:          tags,
		ProjectRefKindMergeRequest: mergeRequests,
	} {
		for _, ref := range refs {
			if _, ok := foundRefs[ref]; ok {
				log.Warn("found duplicate ref for project")
			}
			foundRefs[ref] = kind
		}
	}
	return foundRefs, nil
}

func (c *Client) getProjectBranches(projectID int, refsRegexp string) ([]string, error) {
	var names []string

	options := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	re, err := regexp.Compile(refsRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		branches, resp, err := c.Branches.ListBranches(projectID, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			if re.MatchString(branch.Name) {
				names = append(names, branch.Name)
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *Client) getProjectTags(projectID int, refsRegexp string) ([]string, error) {
	var names []string

	options := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	re, err := regexp.Compile(refsRegexp)
	if err != nil {
		return nil, err
	}

	for {
		c.rateLimit()
		tags, resp, err := c.Tags.ListTags(projectID, options)
		if err != nil {
			return names, err
		}

		for _, tag := range tags {
			if re.MatchString(tag.Name) {
				names = append(names, tag.Name)
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *Client) getProjectMergeRequestsPipelines(projectID int, fetchLimit int) ([]string, error) {
	var names []string

	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	re := regexp.MustCompile(mergeRequestRefRegexp)

	for {
		c.rateLimit()
		pipelines, resp, err := c.Pipelines.ListProjectPipelines(projectID, options)
		if err != nil {
			return nil, fmt.Errorf("error listing project pipelines for project ID %d: %s", projectID, err.Error())
		}

		for _, pipeline := range pipelines {
			if re.MatchString(pipeline.Ref) {
				names = append(names, pipeline.Ref)
				if len(names) >= fetchLimit {
					return names, nil
				}
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *Client) getPipelineVariablesAsConcatenatedString(pr *ProjectRef) (string, error) {
	log.WithFields(
		log.Fields{
			"project-path-with-namespace": pr.PathWithNamespace,
			"project-id":                  pr.ID,
			"pipeline-id":                 pr.MostRecentPipeline.ID,
		},
	).Debug("fetching pipeline variables")

	variablesFilter, err := regexp.Compile(pr.PipelineVariablesRegexp(c.Config))
	if err != nil {
		return "", fmt.Errorf("the provided filter regex for pipeline variables is invalid '(%s)': %v", pr.PipelineVariablesRegexp(c.Config), err)
	}

	c.rateLimit()
	variables, _, err := c.Pipelines.GetPipelineVariables(pr.ID, pr.MostRecentPipeline.ID)
	if err != nil {
		return "", fmt.Errorf("could not fetch pipeline variables for %d: %s", pr.MostRecentPipeline.ID, err.Error())
	}

	var keptVariables []string
	if len(variables) > 0 {
		for _, v := range variables {
			if variablesFilter.MatchString(v.Key) {
				keptVariables = append(keptVariables, strings.Join([]string{v.Key, v.Value}, ":"))
			}
		}
	}

	return strings.Join(keptVariables, ","), nil
}
