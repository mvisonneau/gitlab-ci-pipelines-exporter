package exporter

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func pullProjectRefMetrics(pr schemas.ProjectRef) error {
	logFields := log.Fields{
		"project-name":     pr.PathWithNamespace,
		"project-id":       pr.ID,
		"project-ref":      pr.Ref,
		"project-ref-kind": pr.Kind,
	}

	// TODO: Figure out if we want to have a similar approach for ProjectRefKindTag with
	// an additional configuration parameeter perhaps
	if pr.Kind == schemas.ProjectRefKindMergeRequest && pr.MostRecentPipeline != nil {
		switch pr.MostRecentPipeline.Status {
		case "success", "failed", "canceled", "skipped":
			// The pipeline will not evolve, lets not bother querying the API
			log.WithFields(logFields).WithField("most-recent-pipeline-id", pr.MostRecentPipeline.ID).Debug("skipping finished merge-request pipeline")
			return nil
		}
	}

	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()
	pipelines, err := gitlabClient.GetProjectPipelines(pr.ID, &goGitlab.ListProjectPipelinesOptions{
		// We only need the most recent pipeline
		ListOptions: goGitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
		Ref: goGitlab.String(pr.Ref),
	})

	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", pr.PathWithNamespace, err)
	}

	if len(pipelines) == 0 {
		log.WithFields(logFields).Debug("could not find any pipeline for the project ref")
		return nil
	}

	pipeline, err := gitlabClient.GetProjectRefPipeline(pr, pipelines[0].ID)
	if err != nil {
		return err
	}

	if pr.MostRecentPipeline == nil || !reflect.DeepEqual(pipeline, pr.MostRecentPipeline) {
		pr.MostRecentPipeline = pipeline

		// fetch pipeline variables
		if pr.Pull.Pipeline.Variables.Enabled() {
			pr.MostRecentPipelineVariables, err = gitlabClient.GetProjectRefPipelineVariablesAsConcatenatedString(pr)
			if err != nil {
				return err
			}
		}

		if pipeline.Status == "running" {
			runCount := schemas.Metric{
				Kind:   schemas.MetricKindRunCount,
				Labels: pr.DefaultLabelsValues(),
			}
			storeGetMetric(&runCount)
			runCount.Value++
			storeSetMetric(runCount)
		}

		if pipeline.Coverage != "" {
			parsedCoverage, err := strconv.ParseFloat(pipeline.Coverage, 64)
			if err != nil {
				log.WithFields(logFields).WithField("error", err.Error()).Warnf("could not parse coverage string returned from GitLab API '%s' into Float64", pipeline.Coverage)
			} else {
				storeSetMetric(schemas.Metric{
					Kind:   schemas.MetricKindCoverage,
					Labels: pr.DefaultLabelsValues(),
					Value:  parsedCoverage,
				})
			}
		}

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindID,
			Labels: pr.DefaultLabelsValues(),
			Value:  float64(pipeline.ID),
		})

		emitStatusMetric(
			schemas.MetricKindStatus,
			pr.DefaultLabelsValues(),
			statusesList[:],
			pipeline.Status,
			pr.OutputSparseStatusMetrics(),
		)

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindDurationSeconds,
			Labels: pr.DefaultLabelsValues(),
			Value:  float64(pipeline.Duration),
		})

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindTimestamp,
			Labels: pr.DefaultLabelsValues(),
			Value:  float64(pipeline.UpdatedAt.Unix()),
		})

		if pr.Pull.Pipeline.Jobs.Enabled() {
			if err = pullProjectRefPipelineJobsMetrics(pr); err != nil {
				return err
			}
		}
	}

	if pr.Pull.Pipeline.Jobs.Enabled() {
		if err := pullProjectRefMostRecentJobsMetrics(pr); err != nil {
			return err
		}
	}

	return nil
}
