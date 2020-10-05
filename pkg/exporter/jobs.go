package exporter

import (
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func pollProjectRefPipelineJobs(pr *schemas.ProjectRef) error {
	jobs, err := gitlabClient.ListProjectRefPipelineJobs(pr)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(pr, job)
	}

	return nil
}

func pollProjectRefMostRecentJobs(pr *schemas.ProjectRef) error {
	if !pr.FetchPipelineJobMetrics() {
		return nil
	}

	jobs, err := gitlabClient.ListProjectRefMostRecentJobs(pr)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(pr, job)
	}

	return nil
}

func processJobMetrics(pr *schemas.ProjectRef, job *goGitlab.Job) {
	labels := pr.DefaultLabelsValues()
	labels["stage"] = job.Stage
	labels["job_name"] = job.Name

	// In case a job gets restarted, it will have an ID greated than the previous one(s)
	// jobs in new pipelines should get greated IDs too
	if lastJob, ok := pr.Jobs[job.Name]; ok {
		if lastJob.ID == job.ID {
			store.SetMetric(schemas.Metric{
				Kind:   schemas.MetricKindTimeSinceLastRun,
				Labels: labels,
				Value:  time.Since(*job.CreatedAt).Round(time.Second).Seconds(),
			})
			return
		}
	}

	// Update the job in memory
	pr.Jobs[job.Name] = job
	store.SetProjectRef(*pr)

	log.WithFields(
		log.Fields{
			"project-id":  pr.ID,
			"pipeline-id": pr.MostRecentPipeline.ID,
			"job-name":    job.Name,
			"job-id":      job.ID,
		},
	).Debug("processing job metrics")

	store.SetMetric(schemas.Metric{
		Kind:   schemas.MetricKindLastJobRunID,
		Labels: labels,
		Value:  float64(job.ID),
	})

	store.SetMetric(schemas.Metric{
		Kind:   schemas.MetricKindTimeSinceLastJobRun,
		Labels: labels,
		Value:  time.Since(*job.CreatedAt).Round(time.Second).Seconds(),
	})

	store.SetMetric(schemas.Metric{
		Kind:   schemas.MetricKindLastRunJobDuration,
		Labels: labels,
		Value:  job.Duration,
	})

	emitStatusMetric(
		schemas.MetricKindLastRunJobStatus,
		labels,
		statusesList[:],
		job.Status,
		pr.OutputSparseStatusMetrics(),
	)

	store.SetMetric(schemas.Metric{
		Kind:   schemas.MetricKindTimeSinceLastRun,
		Labels: labels,
		Value:  time.Since(*job.CreatedAt).Round(time.Second).Seconds(),
	})

	jobRunCount := schemas.Metric{
		Kind:   schemas.MetricKindJobRunCount,
		Labels: labels,
	}
	store.PullMetricValue(&jobRunCount)
	jobRunCount.Value++
	store.SetMetric(jobRunCount)

	artifactSize := 0
	for _, artifact := range job.Artifacts {
		artifactSize += artifact.Size
	}

	store.SetMetric(schemas.Metric{
		Kind:   schemas.MetricKindLastRunJobArtifactSize,
		Labels: labels,
		Value:  float64(artifactSize),
	})
}
