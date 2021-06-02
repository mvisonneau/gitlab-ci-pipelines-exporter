package controller

import (
	"reflect"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// PullRefPipelineJobsMetrics ..
func (c *Controller) PullRefPipelineJobsMetrics(ref schemas.Ref) error {
	jobs, err := c.Gitlab.ListRefPipelineJobs(ref)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		c.ProcessJobMetrics(ref, job)
	}

	return nil
}

// PullRefMostRecentJobsMetrics ..
func (c *Controller) PullRefMostRecentJobsMetrics(ref schemas.Ref) error {
	if !ref.PullPipelineJobsEnabled {
		return nil
	}

	jobs, err := c.Gitlab.ListRefMostRecentJobs(ref)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		c.ProcessJobMetrics(ref, job)
	}

	return nil
}

// ProcessJobMetrics ..
func (c *Controller) ProcessJobMetrics(ref schemas.Ref, job schemas.Job) {
	projectRefLogFields := log.Fields{
		"project-name": ref.ProjectName,
		"job-name":     job.Name,
		"job-id":       job.ID,
	}

	labels := ref.DefaultLabelsValues()
	labels["stage"] = job.Stage
	labels["job_name"] = job.Name

	if ref.PullPipelineJobsRunnerDescriptionEnabled {
		re, err := regexp.Compile(ref.PullPipelineJobsRunnerDescriptionAggregationRegexp)
		if err != nil {
			log.WithFields(projectRefLogFields).WithField("error", err.Error()).Error("invalid job runner description aggregation regexp")
		}

		if re.MatchString(job.Runner.Description) {
			labels["runner_description"] = ref.PullPipelineJobsRunnerDescriptionAggregationRegexp
		} else {
			labels["runner_description"] = job.Runner.Description
		}
	} else {
		// TODO: Figure out how to completely remove it from the exporter instead of keeping it empty
		labels["runner_description"] = ""
	}

	// Refresh ref state from the store
	if err := c.Store.GetRef(&ref); err != nil {
		log.WithFields(projectRefLogFields).WithField("error", err.Error()).Error("getting ref from the store")
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
	if err := c.Store.SetRef(ref); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("writing ref in the store")
		return
	}

	log.WithFields(projectRefLogFields).Debug("processing job metrics")

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindJobID,
		Labels: labels,
		Value:  float64(job.ID),
	})

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindJobTimestamp,
		Labels: labels,
		Value:  job.Timestamp,
	})

	storeSetMetric(c.Store, schemas.Metric{
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
	jobRunCountExists, err := c.Store.MetricExists(jobRunCount.Key())
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
		storeGetMetric(c.Store, &jobRunCount)
		jobRunCount.Value++
	}

	storeSetMetric(c.Store, jobRunCount)

	storeSetMetric(c.Store, schemas.Metric{
		Kind:   schemas.MetricKindJobArtifactSizeBytes,
		Labels: labels,
		Value:  job.ArtifactSize,
	})

	emitStatusMetric(
		c.Store,
		schemas.MetricKindJobStatus,
		labels,
		statusesList[:],
		job.Status,
		ref.OutputSparseStatusMetrics,
	)
}
