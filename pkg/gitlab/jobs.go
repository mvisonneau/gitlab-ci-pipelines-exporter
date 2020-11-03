package gitlab

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// ListRefPipelineJobs ..
func (c *Client) ListRefPipelineJobs(ref schemas.Ref) (jobs []schemas.Job, err error) {
	var foundJobs []*goGitlab.Job
	var resp *goGitlab.Response

	if ref.LatestPipeline == (schemas.Pipeline{}) {
		log.WithFields(
			log.Fields{
				"project-name": ref.ProjectName,
				"ref":          ref.Name,
			},
		).Debug("most recent pipeline not defined, exiting..")
		return
	}

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		c.rateLimit()
		foundJobs, resp, err = c.Jobs.ListPipelineJobs(ref.ProjectName, ref.LatestPipeline.ID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			jobs = append(jobs, schemas.NewJob(*job))
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-name": ref.ProjectName,
					"ref":          ref.Name,
					"pipeline-id":  ref.LatestPipeline.ID,
					"jobs-count":   resp.TotalItems,
				},
			).Debug("found pipeline jobs")
			break
		}

		options.Page = resp.NextPage
	}
	return
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

	var foundJobs []goGitlab.Job
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
					jobs = append(jobs, schemas.NewJob(job))
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
			log.WithFields(
				log.Fields{
					"project-name": ref.ProjectName,
					"ref":          ref.Name,
					"jobs-count":   resp.TotalItems,
				},
			).Warn("found some ref jobs but did not manage to refresh all jobs which were in memory")
			break
		}

		options.Page = resp.NextPage
	}
	return
}
