package gitlab

import (
	"strings"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// ListRefPipelineJobs ..
func (c *Client) ListRefPipelineJobs(ref schemas.Ref) (jobs []schemas.Job, err error) {
	if ref.LatestPipeline == (schemas.Pipeline{}) {
		log.WithFields(
			log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting..")
		return
	}

	jobs, err = c.ListPipelineJobs(ref.ProjectName, ref.LatestPipeline.ID)
	if err != nil {
		return
	}

	if ref.PullPipelineJobsFromChildPipelinesEnabled {
		var childJobs []schemas.Job
		childJobs, err = c.ListPipelineChildJobs(ref.ProjectName, ref.LatestPipeline.ID)
		if err != nil {
			return
		}

		jobs = append(jobs, childJobs...)
	}

	return
}

// ListPipelineJobs ..
func (c *Client) ListPipelineJobs(projectName string, pipelineID int) (jobs []schemas.Job, err error) {
	var foundJobs []*goGitlab.Job
	var resp *goGitlab.Response

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit()
		foundJobs, resp, err = c.Jobs.ListPipelineJobs(projectName, pipelineID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			jobs = append(jobs, schemas.NewJob(*job))
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-name": projectName,
					"pipeline-id":  pipelineID,
					"jobs-count":   resp.TotalItems,
				},
			).Debug("found pipeline jobs")
			break
		}

		options.Page = resp.NextPage
	}
	return
}

// ListPipelineBridges ..
func (c *Client) ListPipelineBridges(projectName string, pipelineID int) (bridges []*goGitlab.Bridge, err error) {
	var foundBridges []*goGitlab.Bridge
	var resp *goGitlab.Response

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit()
		foundBridges, resp, err = c.Jobs.ListPipelineBridges(projectName, pipelineID, options)
		if err != nil {
			return
		}

		bridges = append(bridges, foundBridges...)

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-name":  projectName,
					"pipeline-id":   pipelineID,
					"bridges-count": resp.TotalItems,
				},
			).Debug("found pipeline bridges")
			break
		}

		options.Page = resp.NextPage
	}
	return
}

// ListPipelineChildJobs ..
func (c *Client) ListPipelineChildJobs(projectName string, parentPipelineID int) (jobs []schemas.Job, err error) {
	pipelineIDs := []int{parentPipelineID}

	for {
		if len(pipelineIDs) == 0 {
			return
		}

		pipelineID := pipelineIDs[len(pipelineIDs)-1]
		pipelineIDs = pipelineIDs[:len(pipelineIDs)-1]

		var foundBridges []*goGitlab.Bridge
		foundBridges, err = c.ListPipelineBridges(projectName, pipelineID)
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

			pipelineIDs = append(pipelineIDs, foundBridge.DownstreamPipeline.ID)
			var foundJobs []schemas.Job
			foundJobs, err = c.ListPipelineJobs(projectName, foundBridge.DownstreamPipeline.ID)
			if err != nil {
				return
			}

			jobs = append(jobs, foundJobs...)
		}
	}
}

// ListRefMostRecentJobs ..
func (c *Client) ListRefMostRecentJobs(ref schemas.Ref) (jobs []schemas.Job, err error) {
	if len(ref.LatestJobs) == 0 {
		log.WithFields(
			log.Fields{
				"project-name": ref.ProjectName,
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

	var foundJobs []*goGitlab.Job
	var resp *goGitlab.Response

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit()
		foundJobs, resp, err = c.Jobs.ListProjectJobs(ref.ProjectName, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			if _, ok := jobsToRefresh[job.Name]; ok {
				if ref.Name == job.Ref {
					jobs = append(jobs, schemas.NewJob(*job))
					delete(jobsToRefresh, job.Name)
				}
			}

			if len(jobsToRefresh) == 0 {
				log.WithFields(
					log.Fields{
						"project-name": ref.ProjectName,
						"ref":          ref.Name,
						"jobs-count":   len(ref.LatestJobs),
					},
				).Debug("found all jobs to refresh")
				return
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			var notFoundJobs []string
			for k := range jobsToRefresh {
				notFoundJobs = append(notFoundJobs, k)
			}

			log.WithFields(
				log.Fields{
					"project-name":   ref.ProjectName,
					"ref":            ref.Name,
					"jobs-count":     resp.TotalItems,
					"not-found-jobs": strings.Join(notFoundJobs, ","),
				},
			).Warn("found some ref jobs but did not manage to refresh all jobs which were in memory")
			break
		}

		options.Page = resp.NextPage
	}
	return
}
