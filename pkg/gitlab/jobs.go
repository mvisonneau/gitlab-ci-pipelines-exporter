package gitlab

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// ListRefPipelineJobs ..
func (c *Client) ListRefPipelineJobs(ctx context.Context, ref schemas.Ref) (jobs []schemas.Job, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListRefPipelineJobs")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", ref.Project.Name))
	span.SetAttributes(attribute.String("ref_name", ref.Name))

	if reflect.DeepEqual(ref.LatestPipeline, (schemas.Pipeline{})) {
		log.WithFields(
			log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting..")

		return
	}

	jobs, err = c.ListPipelineJobs(ctx, ref.Project.Name, ref.LatestPipeline.ID)
	if err != nil {
		return
	}

	if ref.Project.Pull.Pipeline.Jobs.FromChildPipelines.Enabled {
		var childJobs []schemas.Job

		childJobs, err = c.ListPipelineChildJobs(ctx, ref.Project.Name, ref.LatestPipeline.ID)
		if err != nil {
			return
		}

		jobs = append(jobs, childJobs...)
	}

	return
}

// ListPipelineJobs ..
func (c *Client) ListPipelineJobs(ctx context.Context, projectNameOrID string, pipelineID int) (jobs []schemas.Job, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListPipelineJobs")
	defer span.End()
	span.SetAttributes(attribute.String("project_name_or_id", projectNameOrID))
	span.SetAttributes(attribute.Int("pipeline_id", pipelineID))

	var (
		foundJobs []*goGitlab.Job
		resp      *goGitlab.Response
	)

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit(ctx)

		foundJobs, resp, err = c.Jobs.ListPipelineJobs(projectNameOrID, pipelineID, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		for _, job := range foundJobs {
			jobs = append(jobs, schemas.NewJob(*job))
		}

		if resp.CurrentPage >= resp.NextPage {
			log.WithFields(
				log.Fields{
					"project-name-or-id": projectNameOrID,
					"pipeline-id":        pipelineID,
					"jobs-count":         resp.TotalItems,
				},
			).Debug("found pipeline jobs")

			break
		}

		options.Page = resp.NextPage
	}

	return
}

// ListPipelineBridges ..
func (c *Client) ListPipelineBridges(ctx context.Context, projectNameOrID string, pipelineID int) (bridges []*goGitlab.Bridge, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListPipelineBridges")
	defer span.End()
	span.SetAttributes(attribute.String("project_name_or_id", projectNameOrID))
	span.SetAttributes(attribute.Int("pipeline_id", pipelineID))

	var (
		foundBridges []*goGitlab.Bridge
		resp         *goGitlab.Response
	)

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit(ctx)

		foundBridges, resp, err = c.Jobs.ListPipelineBridges(projectNameOrID, pipelineID, options, goGitlab.WithContext(ctx))
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		bridges = append(bridges, foundBridges...)

		if resp.CurrentPage >= resp.NextPage {
			log.WithFields(
				log.Fields{
					"project-name-or-id": projectNameOrID,
					"pipeline-id":        pipelineID,
					"bridges-count":      resp.TotalItems,
				},
			).Debug("found pipeline bridges")

			break
		}

		options.Page = resp.NextPage
	}

	return
}

// ListPipelineChildJobs ..
func (c *Client) ListPipelineChildJobs(ctx context.Context, projectNameOrID string, parentPipelineID int) (jobs []schemas.Job, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListPipelineChildJobs")
	defer span.End()
	span.SetAttributes(attribute.String("project_name_or_id", projectNameOrID))
	span.SetAttributes(attribute.Int("parent_pipeline_id", parentPipelineID))

	type pipelineDef struct {
		projectNameOrID string
		pipelineID      int
	}

	pipelines := []pipelineDef{{projectNameOrID, parentPipelineID}}

	for {
		if len(pipelines) == 0 {
			return
		}

		var (
			foundBridges []*goGitlab.Bridge
			pipeline     = pipelines[len(pipelines)-1]
		)

		pipelines = pipelines[:len(pipelines)-1]

		foundBridges, err = c.ListPipelineBridges(ctx, pipeline.projectNameOrID, pipeline.pipelineID)
		if err != nil {
			return
		}

		for _, foundBridge := range foundBridges {
			// Trigger job was created but not yet executed
			// so downstream pipeline is not yet scheduled to start.
			// Therefore no pipeline is available and bridge could be skipped.
			if foundBridge.DownstreamPipeline == nil {
				continue
			}

			pipelines = append(pipelines, pipelineDef{strconv.Itoa(foundBridge.DownstreamPipeline.ProjectID), foundBridge.DownstreamPipeline.ID})

			var foundJobs []schemas.Job

			foundJobs, err = c.ListPipelineJobs(ctx, strconv.Itoa(foundBridge.DownstreamPipeline.ProjectID), foundBridge.DownstreamPipeline.ID)
			if err != nil {
				return
			}

			jobs = append(jobs, foundJobs...)
		}
	}
}

// ListRefMostRecentJobs ..
func (c *Client) ListRefMostRecentJobs(ctx context.Context, ref schemas.Ref) (jobs []schemas.Job, err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "gitlab:ListRefMostRecentJobs")
	defer span.End()
	span.SetAttributes(attribute.String("project_name", ref.Project.Name))
	span.SetAttributes(attribute.String("ref_name", ref.Name))

	if len(ref.LatestJobs) == 0 {
		log.WithFields(
			log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
			},
		).Debug("no jobs are currently held in memory, exiting..")

		return
	}

	// Deep copy of the ref.Jobs
	jobsToRefresh := make(schemas.Jobs)
	for k, v := range ref.LatestJobs {
		jobsToRefresh[k] = v
	}

	var (
		foundJobs []*goGitlab.Job
		resp      *goGitlab.Response
		opt       *goGitlab.ListJobsOptions
	)

	keysetPagination := c.Version().PipelineJobsKeysetPaginationSupported()
	if keysetPagination {
		opt = &goGitlab.ListJobsOptions{
			ListOptions: goGitlab.ListOptions{
				Pagination: "keyset",
				PerPage:    100,
			},
		}
	} else {
		opt = &goGitlab.ListJobsOptions{
			ListOptions: goGitlab.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
	}

	options := []goGitlab.RequestOptionFunc{goGitlab.WithContext(ctx)}

	for {
		c.rateLimit(ctx)

		foundJobs, resp, err = c.Jobs.ListProjectJobs(ref.Project.Name, opt, options...)
		if err != nil {
			return
		}

		c.requestsRemaining(resp)

		for _, job := range foundJobs {
			if _, ok := jobsToRefresh[job.Name]; ok {
				jobRefName, _ := schemas.GetMergeRequestIIDFromRefName(job.Ref)
				if ref.Name == jobRefName {
					jobs = append(jobs, schemas.NewJob(*job))
					delete(jobsToRefresh, job.Name)
				}
			}

			if len(jobsToRefresh) == 0 {
				log.WithFields(
					log.Fields{
						"project-name": ref.Project.Name,
						"ref":          ref.Name,
						"jobs-count":   len(ref.LatestJobs),
					},
				).Debug("found all jobs to refresh")

				return
			}
		}

		if keysetPagination && resp.NextLink == "" ||
			(!keysetPagination && resp.CurrentPage >= resp.NextPage) {
			var notFoundJobs []string

			for k := range jobsToRefresh {
				notFoundJobs = append(notFoundJobs, k)
			}

			log.WithContext(ctx).
				WithFields(
					log.Fields{
						"project-name":   ref.Project.Name,
						"ref":            ref.Name,
						"jobs-count":     resp.TotalItems,
						"not-found-jobs": strings.Join(notFoundJobs, ","),
					},
				).
				Warn("found some ref jobs but did not manage to refresh all jobs which were in memory")

			break
		}

		if keysetPagination {
			options = []goGitlab.RequestOptionFunc{
				goGitlab.WithContext(ctx),
				goGitlab.WithKeysetPaginationParameters(resp.NextLink),
			}
		}
	}

	return
}
