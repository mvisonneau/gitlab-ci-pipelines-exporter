package exporter

import (
	"fmt"
	"reflect"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

func pullRefMetrics(ref schemas.Ref) error {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	// At scale, the scheduled ref may be behind the actual state being stored
	// to avoid issues, we refresh it from the store before manipulating it
	if err := store.GetRef(&ref); err != nil {
		return err
	}

	logFields := log.Fields{
		"project-name": ref.ProjectName,
		"ref":          ref.Name,
		"ref-kind":     ref.Kind,
	}

	// TODO: Figure out if we want to have a similar approach for RefKindTag with
	// an additional configuration parameter perhaps
	if ref.Kind == schemas.RefKindMergeRequest && ref.LatestPipeline.ID != 0 {
		switch ref.LatestPipeline.Status {
		case "success", "failed", "canceled", "skipped":
			// The pipeline will not evolve, lets not bother querying the API
			log.WithFields(logFields).WithField("most-recent-pipeline-id", ref.LatestPipeline.ID).Debug("skipping finished merge-request pipeline")
			return nil
		}
	}

	pipelines, err := gitlabClient.GetProjectPipelines(ref.ProjectName, &goGitlab.ListProjectPipelinesOptions{
		// We only need the most recent pipeline
		ListOptions: goGitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
		Ref: goGitlab.String(ref.Name),
	})
	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", ref.ProjectName, err)
	}

	if len(pipelines) == 0 {
		log.WithFields(logFields).Debug("could not find any pipeline for the ref")
		return nil
	}

	pipeline, err := gitlabClient.GetRefPipeline(ref, pipelines[0].ID)
	if err != nil {
		return err
	}

	if ref.LatestPipeline.ID == 0 || !reflect.DeepEqual(pipeline, ref.LatestPipeline) {
		formerPipeline := ref.LatestPipeline
		ref.LatestPipeline = pipeline

		// fetch pipeline variables
		if ref.PullPipelineVariablesEnabled {
			ref.LatestPipeline.Variables, err = gitlabClient.GetRefPipelineVariablesAsConcatenatedString(ref)
			if err != nil {
				return err
			}
		}

		// Update the ref in the store
		if err = store.SetRef(ref); err != nil {
			return err
		}

		// If the metric does not exist yet, start with 0 instead of 1
		// this could cause some false positives in prometheus
		// when restarting the exporter otherwise
		runCount := schemas.Metric{
			Kind:   schemas.MetricKindRunCount,
			Labels: ref.DefaultLabelsValues(),
		}
		storeGetMetric(&runCount)
		if formerPipeline.ID != 0 {
			runCount.Value++
		}
		storeSetMetric(runCount)

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindCoverage,
			Labels: ref.DefaultLabelsValues(),
			Value:  pipeline.Coverage,
		})

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindID,
			Labels: ref.DefaultLabelsValues(),
			Value:  float64(pipeline.ID),
		})

		emitStatusMetric(
			schemas.MetricKindStatus,
			ref.DefaultLabelsValues(),
			statusesList[:],
			pipeline.Status,
			ref.OutputSparseStatusMetrics,
		)

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindDurationSeconds,
			Labels: ref.DefaultLabelsValues(),
			Value:  pipeline.DurationSeconds,
		})

		storeSetMetric(schemas.Metric{
			Kind:   schemas.MetricKindTimestamp,
			Labels: ref.DefaultLabelsValues(),
			Value:  pipeline.Timestamp,
		})

		if ref.PullPipelineJobsEnabled {
			if err := pullRefPipelineJobsMetrics(ref); err != nil {
				return err
			}
		}
		return nil
	}

	if ref.PullPipelineJobsEnabled {
		if err := pullRefMostRecentJobsMetrics(ref); err != nil {
			return err
		}
	}

	return nil
}
