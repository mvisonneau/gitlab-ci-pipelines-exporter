package exporter

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"

	log "github.com/sirupsen/logrus"
)

func pullRefPipelineJobsMetrics(ref schemas.Ref, pullJobTraces bool) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	jobs, err := gitlabClient.ListRefPipelineJobs(ref)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		processJobMetrics(ref, job, pullJobTraces)
	}

	return nil
}

func pullRefMostRecentJobsMetrics(ref schemas.Ref, pullJobTraces bool) error {
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
		processJobMetrics(ref, job, pullJobTraces)
	}

	return nil
}

// ParseJobTrace ..
// func (c *Client) ParseJobTrace(projectName string, jobID int, rules []string) (parsedOutput string, err error) {

// 	trace, _, err := c.Jobs.GetTraceFile(projectName, jobID)

// 	buf := new(bytes.Buffer)
// 	buf.ReadFrom(trace)
// 	parsedOutput = buf.String()

// 	r, _ := regexp.Compile("---> (.*)")
// 	foundStrings := r.FindAllString(parsedOutput, -1)

// 	for i, s := range foundStrings {
// 		fmt.Println(i, s)
// 	}

// 	return

// }

func processJobTrace(ref schemas.Ref, job schemas.Job) {
	fmt.Println("Pull and parse trace output for: ", ref.ProjectName, " / ", job.Name, " / ", job.ID)

	var newTraceMatch schemas.TraceMatch
	var foundTraceMatches []schemas.TraceMatch

	for _, configRule := range config.Pull.TraceRules {
		for _, jobRule := range ref.PullPipelineJobsTraceRules {
			if configRule.Name == jobRule {
				fmt.Println("This job trace process we must: ", configRule.RegexpValue)
				newTraceMatch.RuleName = configRule.Name
				newTraceMatch.RegexpValue = configRule.RegexpValue
				foundTraceMatches = append(foundTraceMatches, newTraceMatch)
			}
		}
	}

	fmt.Println(foundTraceMatches)

	trace, _, err := gitlabClient.Jobs.GetTraceFile(ref.ProjectName, job.ID)

	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(trace)
	parsedOutput := buf.String()

	r, _ := regexp.Compile("---> (.*)")
	foundStrings := r.FindAllString(parsedOutput, -1)

	for i, s := range foundStrings {
		fmt.Println(i, s)
	}

}

func processJobMetrics(ref schemas.Ref, job schemas.Job, pullJobTraces bool) {

	// Trace match metrics
	if pullJobTraces && len(ref.PullPipelineJobsTraceRules) > 0 {
		processJobTrace(ref, job)
	}

	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

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
	if err := store.GetRef(&ref); err != nil {
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
	if err := store.SetRef(ref); err != nil {
		log.WithFields(
			projectRefLogFields,
		).WithField("error", err.Error()).Error("writing ref in the store")
		return
	}

	log.WithFields(projectRefLogFields).Debug("processing job metrics")

	// Trace match metrics
	if pullJobTraces && len(ref.PullPipelineJobsTraceRules) > 0 {
		processJobTrace(ref, job)
		labels["trace_rule"] = ""
	} else {
		labels["trace_rule"] = ""
	}

	storeSetMetric(schemas.Metric{
		Kind:   schemas.MetricKindJobTraceMatchCount,
		Labels: labels,
		Value:  123,
	})

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
