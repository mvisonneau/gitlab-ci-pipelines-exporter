package gitlab

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// ListProjectRefPipelineJobs ..
func (c *Client) ListProjectRefPipelineJobs(pr *schemas.ProjectRef) (jobs []*goGitlab.Job, err error) {
	var foundJobs []*goGitlab.Job
	var resp *goGitlab.Response

	// Initialize the variable
	if pr.Jobs == nil {
		pr.Jobs = map[string]*goGitlab.Job{}
	}

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		c.rateLimit()
		foundJobs, resp, err = c.Jobs.ListPipelineJobs(pr.ID, pr.MostRecentPipeline.ID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			jobs = append(jobs, job)
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-id":  pr.ID,
					"project-ref": pr.Ref,
					"pipeline-id": pr.MostRecentPipeline.ID,
					"jobs-count":  resp.TotalItems,
				},
			).Info("found pipeline jobs")
			break
		}

		options.Page = resp.NextPage
	}
	return
}

// ListProjectRefMostRecentJobs ..
func (c *Client) ListProjectRefMostRecentJobs(pr *schemas.ProjectRef) (jobs []*goGitlab.Job, err error) {
	if pr.Jobs == nil {
		log.WithFields(
			log.Fields{
				"project-id":  pr.ID,
				"project-ref": pr.Ref,
			},
		).Debug("no jobs are currently held in memory, exiting..")
		return
	}

	jobsToRefresh := pr.Jobs

	var foundJobs []goGitlab.Job
	var resp *goGitlab.Response

	options := &goGitlab.ListJobsOptions{
		ListOptions: goGitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		c.rateLimit()
		foundJobs, resp, err = c.Jobs.ListProjectJobs(pr.ID, options)
		if err != nil {
			return
		}

		for _, job := range foundJobs {
			if jobToRefresh, ok := jobsToRefresh[job.Name]; ok {
				if jobToRefresh.Ref == job.Ref {
					jobs = append(jobs, &job)
					delete(jobsToRefresh, job.Name)
				}
			}

			if len(jobsToRefresh) == 0 {
				log.WithFields(
					log.Fields{
						"project-id":  pr.ID,
						"project-ref": pr.Ref,
						"jobs-count":  len(pr.Jobs),
					},
				).Info("found all jobs to refresh")
				return
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			log.WithFields(
				log.Fields{
					"project-id":  pr.ID,
					"project-ref": pr.Ref,
					"jobs-count":  resp.TotalItems,
				},
			).Warn("found some project ref jobs but did not manage to refresh all jobs which were in memory")
			break
		}

		options.Page = resp.NextPage
	}
	return
}
