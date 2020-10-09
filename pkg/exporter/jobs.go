package exporter

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func pullProjectRefPipelineJobsMetrics(pr schemas.ProjectRef) error {
	jobs, err := gitlabClient.ListProjectRefPipelineJobs(pr)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(pr, job)
	}

	return nil
}

func pullProjectRefMostRecentJobsMetrics(pr schemas.ProjectRef) error {
	if !pr.Pull.Pipeline.Jobs.Enabled() {
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

func processJobMetrics(pr schemas.ProjectRef, job goGitlab.Job) {
	labels := pr.DefaultLabelsValues()
	labels["stage"] = job.Stage
	labels["job_name"] = job.Name

	projectRefLogFields := log.Fields{
		"project-id": pr.ID,
		"job-name":   job.Name,
		"job-id":     job.ID,
	}

	if err := store.GetProjectRef(&pr); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("getting project ref from the store")
		return
	}

	// In case a job gets restarted, it will have an ID greated than the previous one(s)
	// jobs in new pipelines should get greated IDs too
	if lastJob, ok := pr.Jobs[job.Name]; ok {
		if lastJob.ID > job.ID {
			return
		}
	}

	// Update the project ref in the store
	pr.Jobs[job.Name] = job
	if err := store.SetProjectRef(pr); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("writing project ref in the store")
		return
	}

	log.WithFields(projectRefLogFields).Debug("processing job metrics")

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobID,
		Labels: labels,
		Value:  float64(job.ID),
	})

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobTimestamp,
		Labels: labels,
		Value:  float64(job.CreatedAt.Unix()),
	})

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobDurationSeconds,
		Labels: labels,
		Value:  job.Duration,
	})

	jobRunCount := schemas.Metric{
		Kind:   schemas.MetricKindJobRunCount,
		Labels: labels,
	}
	storeGetMetric(&jobRunCount)
	jobRunCount.Value++
	storeSetMetric(jobRunCount)

	artifactSize := 0
	for _, artifact := range job.Artifacts {
		artifactSize += artifact.Size
	}

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobArtifactSizeBytes,
		Labels: labels,
		Value:  float64(artifactSize),
	})

	emitStatusMetric(
		schemas.MetricKindJobStatus,
		labels,
		statusesList[:],
		job.Status,
		pr.OutputSparseStatusMetrics(),
	)
}
