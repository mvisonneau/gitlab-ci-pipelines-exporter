package gitlab

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// ListRefPipelineJobs ..
func (c *Client) ListRefPipelineJobs(ref schemas.Ref) (jobs []goGitlab.Job, err error) {
	var foundJobs []*goGitlab.Job
	var resp *goGitlab.Response

	if ref.MostRecentPipeline == nil {
		log.WithFields(
			log.Fields{
				"project-id":  ref.ID,
				"project-ref": ref.Ref,
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
		foundJobs, resp, err = c.Jobs.ListPipelineJobs(ref.ID, ref.MostRecentPipeline.ID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			jobs = append(jobs, *job)
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-id":  ref.ID,
					"project-ref": ref.Ref,
					"pipeline-id": ref.MostRecentPipeline.ID,
					"jobs-count":  resp.TotalItems,
				},
			).Info("found pipeline jobs")
			break
		}

		options.Page = resp.NextPage
	}
	return
}

// ListRefMostRecentJobs ..
func (c *Client) ListRefMostRecentJobs(ref schemas.Ref) (jobs []goGitlab.Job, err error) {
	if ref.Jobs == nil {
		log.WithFields(
			log.Fields{
				"project-id":  ref.ID,
				"project-ref": ref.Ref,
			},
		).Debug("no jobs are currently held in memory, exiting..")
		return
	}

	// Deep copy of the ref.Jobs
	jobsToRefresh := make(map[string]goGitlab.Job)
	for k, v := range ref.Jobs {
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
		foundJobs, resp, err = c.Jobs.ListProjectJobs(ref.ID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			if _, ok := jobsToRefresh[job.Name]; ok {
				if ref.Ref == job.Ref {
					jobs = append(jobs, job)
					delete(jobsToRefresh, job.Name)
				}
			}

			if len(jobsToRefresh) == 0 {
				log.WithFields(
					log.Fields{
						"project-id":  ref.ID,
						"project-ref": ref.Ref,
						"jobs-count":  len(ref.Jobs),
					},
				).Info("found all jobs to refresh")
				return
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-id":  ref.ID,
					"project-ref": ref.Ref,
					"jobs-count":  resp.TotalItems,
				},
			).Warn("found some project ref jobs but did not manage to refresh all jobs which were in memory")
			break
		}

		options.Page = resp.NextPage
	}
	return
}
