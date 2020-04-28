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

var statusesList = [...]string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"}

// ProjectRef is what we will use a metrics entity on which we will
// perform regular polling operations
type ProjectRef struct {
	*schemas.Project

	ID                          int
	PathWithNamespace           string
	Topics                      string
	Ref                         string
	MostRecentPipeline          *gitlab.Pipeline
	MostRecentPipelineVariables string
}

// ProjectsRefs allows us to keep track of all the ProjectRef
// we have configured/discovered
type ProjectsRefs map[int]map[string]*ProjectRef

// NewProjectRef is an helper which returns a new ProjectRef pointer
func NewProjectRef(project *schemas.Project, gp *gitlab.Project, ref string) *ProjectRef {
	return &ProjectRef{
		Project:           project,
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
}

func (pr *ProjectRef) defaultLabelsValues() []string {
	return []string{pr.PathWithNamespace, pr.Topics, pr.Ref, pr.MostRecentPipelineVariables}
}

// OrchestratePolling ...
func (c *Client) OrchestratePolling(until <-chan bool) {
	go func() {
		discoverWildcardsProjectsEvery := time.NewTicker(time.Duration(c.Config.WildcardsProjectsDiscoverIntervalSeconds) * time.Second)
		discoverProjectsRefsEvery := time.NewTicker(time.Duration(c.Config.ProjectsRefsDiscoverIntervalSeconds) * time.Second)
		pollProjectsRefsEvery := time.NewTicker(time.Duration(c.Config.ProjectsRefsPollingIntervalSeconds) * time.Second)
		stopWorkers := make(chan struct{})
		defer close(stopWorkers)
		// first execution, blocking call before entering orchestration loop to get the wildcards discovered
		c.discoverProjectsRefsFromPipelinesOnInit()
		c.discoverProjectsFromWildcards()
		c.discoverProjectsRefs()
		c.pollProjectsRefsUntil(stopWorkers)

		for {
			select {
			case <-until:
				log.Info("stopping projects polling...")
				return
			case <-pollProjectsRefsEvery.C:
				// poll all the configured project refs pipelines
				c.pollProjectsRefsUntil(stopWorkers)
			case <-discoverProjectsRefsEvery.C:
				// refresh the list of project refs
				c.discoverProjectsRefs()
			case <-discoverWildcardsProjectsEvery.C:
				// refresh the list of projects from wildcards
				c.discoverProjectsFromWildcards()
			}
		}
	}()
}

func (c *Client) pollProject(p schemas.Project) error {
	log.WithFields(
		log.Fields{
			"project-name": p.Name,
		},
	).Debug("fetching project")

	project, err := c.getProject(p.Name)
	if err != nil {
		return fmt.Errorf("unable to fetch project '%s' from the GitLab API: %s", p.Name, err.Error())
	}

	refs, err := c.getProjectRefs(project.ID, p.RefsRegexp(c.Config))
	if err != nil {
		return fmt.Errorf("error fetching refs for project '%s': %s", p.Name, err.Error())
	}
	if len(refs) == 0 {
		log.WithFields(
			log.Fields{
				"project-name": p.Name,
			},
		).Warn("no refs found for project")
		return nil
	}

	// read the metrics for refs
	log.WithFields(
		log.Fields{
			"project-name": p.Name,
		},
	).Debug("polling project refs")
	for _, ref := range refs {
		pr := NewProjectRef(&p, project, ref)
		log.WithFields(
			log.Fields{
				"project-path-with-namespace": pr.PathWithNamespace,
				"project-ref":                 ref,
			},
		).Info("discovered new project ref")
		if err := c.pollProjectRefMostRecentPipeline(pr); err != nil {
			log.WithFields(
				log.Fields{
					"project-path-with-namespace": pr.PathWithNamespace,
					"project-ref":                 ref,
					"error":                       err.Error(),
				},
			).Error("getting pipeline data for a project ref")
			continue
		}
	}

	return nil
}

func (c *Client) pollPipelineJobs(pr *ProjectRef) error {
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
		jobs, resp, err = c.Jobs.ListPipelineJobs(pr.ID, pr.MostRecentPipeline.ID, options)
		if err != nil {
			return err
		}

		// otherwise proceed
		log.Infof("Found %d jobs for pipeline %d", len(jobs), pr.MostRecentPipeline.ID)
		for _, job := range jobs {
			jobValues := append(pr.defaultLabelsValues(), job.Stage, job.Name)

			log.Debugf("Job %s for pipeline %d", job.Name, pr.MostRecentPipeline.ID)
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

	// TODO: Evaluate if a > operator would not make even more sense here
	if pr.MostRecentPipeline == nil || pipeline.ID != pr.MostRecentPipeline.ID {
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

		defaultLabelValues := pr.defaultLabelsValues()
		runCount.WithLabelValues(defaultLabelValues...).Inc()

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

		if pr.FetchPipelineJobMetrics(c.Config) {
			if err := c.pollPipelineJobs(pr); err != nil {
				log.WithFields(
					log.Fields{
						"project-path-with-namespace": pr.PathWithNamespace,
						"project-id":                  pr.ID,
						"project-ref":                 pr.Ref,
						"pipeline-id":                 pipeline.ID,
						"error":                       err.Error(),
					},
				).Error("polling pipeline jobs")
			}
		}

		timeSinceLastRun.WithLabelValues(defaultLabelValues...).Set(time.Since(*pipeline.CreatedAt).Round(time.Second).Seconds())
	}

	return nil
}

func (c *Client) discoverProjectsFromWildcards() {
	log.WithFields(
		log.Fields{
			"count": len(c.Config.Wildcards),
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
		).Debug("not polling most recent project pipelines")
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

func (c *Client) pollProjectsRefsUntil(stop <-chan struct{}) {
	log.WithFields(
		log.Fields{
			"count": len(c.ProjectsRefs),
		},
	).Info("polling metrics from projects refs")

	pollErrors := pollProjectsRefs(c.Config.MaximumProjectsPollingWorkers, c.pollProjectRefMostRecentPipeline, stop, c.ProjectsRefs)
	for err := range pollErrors {
		if err != nil {
			log.WithFields(
				log.Fields{
					"error": err.Error(),
				},
			).Error("whilst metrics from polling projects refs")
		}
	}
}

func pollProjectsRefs(numWorkers int, pollFunction func(*ProjectRef) error, until <-chan struct{}, projectsRefs ProjectsRefs) <-chan error {
	errorStream := make(chan error)
	projectsRefsToPoll := make(chan *ProjectRef)
	// sync closing the error channel via a waitGroup
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	// spawn maximum_projects_poller_workers to process project polling in parallel
	for w := 0; w < numWorkers; w++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for pr := range projectsRefsToPoll {
				select {
				case <-until:
					return
				case errorStream <- pollFunction(pr):
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
	for _, refs := range projectsRefs {
		for _, pr := range refs {
			projectsRefsToPoll <- pr
		}
	}

	close(projectsRefsToPoll)
	return errorStream
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
	}

	pipelines, err := c.getProjectPipelines(gp.ID, options)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(p.RefsRegexp(c.Config))
	if err != nil {
		return nil, err
	}

	projectRefs := map[string]*ProjectRef{}
	for _, pipeline := range pipelines {
		if re.MatchString(pipeline.Ref) {
			if _, ok := projectRefs[pipeline.Ref]; !ok {
				log.WithFields(
					log.Fields{
						"project-id":                  gp.ID,
						"project-path-with-namespace": gp.PathWithNamespace,
						"project-ref":                 pipeline.Ref,
					},
				).Info("found project ref")
				projectRefs[pipeline.Ref] = NewProjectRef(p, gp, pipeline.Ref)
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
		).Debug("throttling GitLab requests")
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

		refs, err := c.getProjectRefs(gp.ID, p.RefsRegexp(c.Config))
		if err != nil {
			log.WithFields(
				log.Fields{
					"project-id":                  gp.ID,
					"project-path-with-namespace": gp.PathWithNamespace,
					"error":                       err.Error(),
				},
			).Error("getting project refs")
		}

		for _, ref := range refs {
			if _, ok := c.ProjectsRefs[gp.ID][ref]; !ok {
				log.WithFields(
					log.Fields{
						"project-id":                  gp.ID,
						"project-path-with-namespace": gp.PathWithNamespace,
						"project-ref":                 ref,
					},
				).Info("discovered new project ref")
				c.ProjectsRefs[gp.ID][ref] = NewProjectRef(&p, gp, ref)
			}
		}
	}
}

func (c *Client) getProjectRefs(projectID int, refsRegexp string) ([]string, error) {
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
