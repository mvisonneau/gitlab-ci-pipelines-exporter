package exporter

import (
	"reflect"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
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

func processJobMetrics(ref schemas.Ref, job schemas.Job) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	labels := ref.DefaultLabelsValues()
	labels["stage"] = job.Stage
	labels["job_name"] = job.Name
	labels["runner_description"] = job.Runner.Description

	projectRefLogFields := log.Fields{
		"project-name": ref.ProjectName,
		"job-name":     job.Name,
		"job-id":       job.ID,
	}

	// Refresh ref state from the store
	if err := store.GetRef(&ref); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("getting ref from the store")
		return
	}

	// In case a job gets restarted, it will have an ID greated than the previous one(s)
	// jobs in new pipelines should get greated IDs too
	lastJob, lastJobExists := ref.LatestJobs[job.Name]
	if lastJobExists && reflect.DeepEqual(lastJob, job) {
		return
	}

	// Update the ref in the store
	if ref.LatestJobs == nil {
		ref.LatestJobs = make(schemas.Jobs)
	}
	ref.LatestJobs[job.Name] = job
	if err := store.SetRef(ref); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("writing ref in the store")
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
		Value:  job.Timestamp,
	})

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobDurationSeconds,
		Labels: labels,
		Value:  job.DurationSeconds,
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
		).WithField("error", err.Error()).Error("checking if metric exists in the store")
		return
	}

	// We want to increment this counter only once per job ID if:
	// - the metric is already set
	// - the job has been triggered
	jobTriggeredRegexp := regexp.MustCompile("^(skipped|manual|scheduled)$")
	lastJobTriggered := !jobTriggeredRegexp.MatchString(lastJob.Status)
	jobTriggered := !jobTriggeredRegexp.MatchString(job.Status)
	if jobRunCountExists && ((lastJob.ID != job.ID && jobTriggered) || (lastJob.ID == job.ID && jobTriggered && !lastJobTriggered)) {
		storeGetMetric(&jobRunCount)
		jobRunCount.Value++
	}

	storeSetMetric(jobRunCount)

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobArtifactSizeBytes,
		Labels: labels,
		Value:  job.ArtifactSize,
	})

	emitStatusMetric(
		schemas.MetricKindJobStatus,
		labels,
		statusesList[:],
		job.Status,
		ref.OutputSparseStatusMetrics,
	)
}
