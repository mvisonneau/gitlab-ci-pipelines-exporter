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

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
	Config      *schemas.Config
	RateLimiter ratelimit.Limiter
}

type ProjectDetails struct {
	*schemas.Project

	ID                          int
	PathWithNamespace           string
	Topics                      string
	Ref                         string
	MostRecentPipeline          *gitlab.Pipeline
	MostRecentPipelineVariables string
}

func NewProjectDetails(project *schemas.Project, gp *gitlab.Project, ref string) *ProjectDetails {
	return &ProjectDetails{
		Project:           project,
		ID:                gp.ID,
		PathWithNamespace: gp.PathWithNamespace,
		Topics:            strings.Join(gp.TagList, ","),
		Ref:               ref,
	}
}

func (pd *ProjectDetails) defaultLabelsValues() []string {
	return []string{pd.PathWithNamespace, pd.Topics, pd.Ref, pd.MostRecentPipelineVariables}
}

func refExists(r string, refs []string) bool {
	for _, ref := range refs {
		if ref == r {
			return true
		}
	}
	return false
}

// OrchestratePolling ...
func (c *Client) OrchestratePolling(until <-chan bool, getRefsOnInit bool) {
	go func() {
		pollWildcardsEvery := time.NewTicker(time.Duration(c.Config.ProjectsPollingIntervalSeconds) * time.Second)
		pollRefsEvery := time.NewTicker(time.Duration(c.Config.RefsPollingIntervalSeconds) * time.Second)
		stopWorkers := make(chan struct{})
		defer close(stopWorkers)
		// first execution, blocking call before entering orchestration loop to get the wildcards discovered
		c.discoverWildcards()
		c.pollWithWorkersUntil(stopWorkers)

		if getRefsOnInit {
			c.pollPipelinesOnInit()
		}

		for {
			select {
			case <-until:
				log.Info("stopping projects polling...")
				return
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

	branchesAndTagRefs, err := c.branchesAndTagsFor(project.ID, p.RefsRegexp(c.Config))
	if err != nil {
		return fmt.Errorf("error fetching refs for project '%s': %s", p.Name, err.Error())
	}
	if len(branchesAndTagRefs) == 0 {
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
	for _, ref := range branchesAndTagRefs {
		pd := NewProjectDetails(&p, project, ref)
		log.WithFields(
			log.Fields{
				"project-path-with-namespace": pd.PathWithNamespace,
				"project-ref":                 ref,
			},
		).Info("found project refs")
		if err := c.pollProjectRef(pd); err != nil {
			log.WithFields(
				log.Fields{
					"project-path-with-namespace": pd.PathWithNamespace,
					"project-ref":                 ref,
					"error":                       err.Error(),
				},
			).Error("getting pipeline data for a project ref")
			continue
		}
	}

	return nil
}

func (c *Client) pollPipelineJobs(pd *ProjectDetails) error {
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
		jobs, resp, err = c.Jobs.ListPipelineJobs(pd.ID, pd.MostRecentPipeline.ID, options)
		if err != nil {
			return err
		}

		// otherwise proceed
		log.Infof("Found %d jobs for pipeline %d", len(jobs), pd.MostRecentPipeline.ID)
		for _, job := range jobs {
			jobValues := append(pd.defaultLabelsValues(), job.Stage, job.Name)

			log.Debugf("Job %s for pipeline %d", job.Name, pd.MostRecentPipeline.ID)
			lastRunJobDuration.WithLabelValues(jobValues...).Set(job.Duration)

			emitStatusMetric(
				lastRunJobStatus,
				jobValues,
				statusesList[:],
				job.Status,
				pd.OutputSparseStatusMetrics(c.Config),
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

func (c *Client) pollProjectRef(pd *ProjectDetails) error {
	pipelines, err := c.getProjectPipelines(pd, &gitlab.ListProjectPipelinesOptions{
		// We only need the most recent pipeline
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
		Ref: gitlab.String(pd.Ref),
	})

	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", pd.PathWithNamespace, err)
	}

	if len(pipelines) == 0 {
		return fmt.Errorf("could not find any pipeline for project %s with ref %s", pd.PathWithNamespace, pd.Ref)
	}

	c.rateLimit()
	pipeline, _, err := c.Pipelines.GetPipeline(pd.ID, pipelines[0].ID)
	if err != nil || pipeline == nil {
		return fmt.Errorf("could not read content of last pipeline %s:%s", pd.PathWithNamespace, pd.Ref)
	}

	// TODO: Evaluate if a > operator would not make even more sense here
	if pd.MostRecentPipeline == nil || pipeline.ID != pd.MostRecentPipeline.ID {
		pd.MostRecentPipeline = pipeline

		// fetch pipeline variables

		if pd.FetchPipelineVariables(c.Config) {
			pd.MostRecentPipelineVariables, err = c.getPipelineVariablesAsConcatenatedString(pd)
			if err != nil {
				return err
			}
		} else {
			// Ensure we flush the value if there was some variables defined on the previous pipeline
			pd.MostRecentPipelineVariables = ""
		}

		defaultLabelValues := pd.defaultLabelsValues()
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
			pd.OutputSparseStatusMetrics(c.Config),
		)

		if pd.FetchPipelineJobMetrics(c.Config) {
			if err := c.pollPipelineJobs(pd); err != nil {
				log.WithFields(
					log.Fields{
						"project-path-with-namespace": pd.PathWithNamespace,
						"project-id":                  pd.ID,
						"project-ref":                 pd.Ref,
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

func (c *Client) discoverWildcards() {
	log.WithFields(
		log.Fields{
			"count": len(c.Config.Wildcards),
		},
	).Info("configured wildcards")

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
				).Infof("found new project")
				c.Config.Projects = append(c.Config.Projects, p)
			}
		}
	}
}

func (c *Client) pollPipelinesOnInit() {
	log.WithFields(
		log.Fields{
			"init-operation": true,
		},
	).Debug("polling most recent project pipelines")

	for _, p := range c.Config.Projects {
		log.WithFields(
			log.Fields{
				"init-operation": true,
				"project-name":   p.Name,
			},
		).Debug("fetching project")

		gitlabProject, err := c.getProject(p.Name)
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
		// TODO: It would be nice to remove this project details object if not going to
		// be used afterwards for metrics purposes
		pipelineRefs, err := c.refsFromPipelines(NewProjectDetails(&p, gitlabProject, ""))
		if err != nil {
			log.WithFields(
				log.Fields{
					"init-operation": true,
					"project-name":   p.Name,
				},
			).Errorf("unable to fetch refs from project pipelines: %s", err.Error())
			continue
		}
		for _, ref := range pipelineRefs {
			if err := c.pollProjectRef(NewProjectDetails(&p, gitlabProject, ref)); err != nil {
				log.WithFields(
					log.Fields{
						"init-operation": true,
						"project-name":   p.Name,
						"project-ref":    ref,
					},
				).Errorf("unable to poll project ref: %s", err.Error())
			}
		}
	}
}

func (c *Client) pollWithWorkersUntil(stop <-chan struct{}) {
	log.WithFields(
		log.Fields{
			"count": len(c.Config.Projects),
		},
	).Info("configured projects")

	pollErrors := pollProjectsWith(c.Config.MaximumProjectsPollingWorkers, c.pollProject, stop, c.Config.Projects...)
	for err := range pollErrors {
		if err != nil {
			log.WithFields(
				log.Fields{
					"error": err.Error(),
				},
			).Error("whilst polling projects")
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

func (c *Client) getProjectPipelines(pd *ProjectDetails, options *gitlab.ListProjectPipelinesOptions) ([]*gitlab.PipelineInfo, error) {
	log.WithFields(
		log.Fields{
			"project-path-with-namespace": pd.PathWithNamespace,
			"project-id":                  pd.ID,
		},
	).Debug("listing project pipelines")

	c.rateLimit()

	pipelines, _, err := c.Pipelines.ListProjectPipelines(pd.ID, options)
	if err != nil {
		return nil, fmt.Errorf("error listing project pipelines for project %s: %v", pd.PathWithNamespace, err)
	}
	return pipelines, nil
}

func (c *Client) refsFromPipelines(pd *ProjectDetails) ([]string, error) {
	options := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
			// TODO: Get a proper loop to split this query up
			PerPage: c.Config.OnInitFetchRefsFromPipelinesDepthLimit,
		},
	}
	pipelines, err := c.getProjectPipelines(pd, options)
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

func (c *Client) getPipelineVariablesAsConcatenatedString(pd *ProjectDetails) (string, error) {
	log.WithFields(
		log.Fields{
			"project-path-with-namespace": pd.PathWithNamespace,
			"project-id":                  pd.ID,
			"pipeline-id":                 pd.MostRecentPipeline.ID,
		},
	).Debug("fetching pipeline variables")

	variablesFilter, err := regexp.Compile(pd.PipelineVariablesRegexp(c.Config))
	if err != nil {
		return "", fmt.Errorf("the provided filter regex for pipeline variables is invalid '(%s)': %v", pd.PipelineVariablesRegexp(c.Config), err)
	}

	c.rateLimit()
	variables, _, err := c.Pipelines.GetPipelineVariables(pd.ID, pd.MostRecentPipeline.ID)
	if err != nil {
		return "", fmt.Errorf("could not fetch pipeline variables for %d: %s", pd.MostRecentPipeline.ID, err.Error())
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
