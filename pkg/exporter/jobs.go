package exporter

import (
	"reflect"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func pullRefPipelineJobsMetrics(ref schemas.Ref) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	jobs, err := gitlabClient.ListRefPipelineJobs(ref)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(ref, job)
	}

	return nil
}

func pullRefMostRecentJobsMetrics(ref schemas.Ref) error {
	if !ref.PullPipelineJobsEnabled {
		return nil
	}

	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	jobs, err := gitlabClient.ListRefMostRecentJobs(ref)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(ref, job)
	}

	return nil
}

func processJobMetrics(ref schemas.Ref, job goGitlab.Job) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	labels := ref.DefaultLabelsValues()
	labels["stage"] = job.Stage
	labels["job_name"] = job.Name

	projectRefLogFields := log.Fields{
		"project-name": ref.ProjectName,
		"job-name":     job.Name,
		"job-id":       job.ID,
	}

	// Refresh ref state from the store
	if err := store.GetRef(&ref); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("getting project ref from the store")
		return
	}

	// In case a job gets restarted, it will have an ID greated than the previous one(s)
	// jobs in new pipelines should get greated IDs too
	if lastJob, ok := ref.Jobs[job.Name]; ok && reflect.DeepEqual(lastJob, job) {
		return
	}

	// Update the project ref in the store
	ref.Jobs[job.Name] = job
	if err := store.SetRef(ref); err != nil {
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

	// If the metric does not exist yet, start with 0 instead of 1
	// this could cause some false positives in prometheus
	// when restarting the exporter otherwise
	jobRunCountExists, err := store.MetricExists(jobRunCount.Key())
	if err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("getting metric from the store")
		return
	}

	if jobRunCountExists {
		storeGetMetric(&jobRunCount)
		jobRunCount.Value++
	}

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
		ref.OutputSparseStatusMetrics,
	)
}
