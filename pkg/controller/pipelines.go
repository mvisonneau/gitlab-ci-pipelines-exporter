package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	goGitlab "github.com/xanzy/go-gitlab"
)

// PullRefMetrics ..
func (c *Controller) PullRefMetrics(ctx context.Context, ref schemas.Ref) error {
	// At scale, the scheduled ref may be behind the actual state being stored
	// to avoid issues, we refresh it from the store before manipulating it
	if err := c.Store.GetRef(ctx, &ref); err != nil {
		return err
	}

	logFields := log.Fields{
		"project-name": ref.Project.Name,
		"ref":          ref.Name,
		"ref-kind":     ref.Kind,
	}

	// We need a different syntax if the ref is a merge-request
	var refName string
	if ref.Kind == schemas.RefKindMergeRequest {
		refName = fmt.Sprintf("refs/merge-requests/%s/head", ref.Name)
	} else {
		refName = ref.Name
	}

	pipelines, _, err := c.Gitlab.GetProjectPipelines(ctx, ref.Project.Name, &goGitlab.ListProjectPipelinesOptions{
		// We only need the most recent pipeline
		ListOptions: goGitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
		Ref: &refName,
	})
	if err != nil {
		return fmt.Errorf("error fetching project pipelines for %s: %v", ref.Project.Name, err)
	}

	if len(pipelines) == 0 {
		log.WithFields(logFields).Debug("could not find any pipeline for the ref")

		return nil
	}

	pipeline, err := c.Gitlab.GetRefPipeline(ctx, ref, pipelines[0].ID)
	if err != nil {
		return err
	}

	if ref.LatestPipeline.ID == 0 || !reflect.DeepEqual(pipeline, ref.LatestPipeline) {
		formerPipeline := ref.LatestPipeline
		ref.LatestPipeline = pipeline

		// fetch pipeline variables
		if ref.Project.Pull.Pipeline.Variables.Enabled {
			ref.LatestPipeline.Variables, err = c.Gitlab.GetRefPipelineVariablesAsConcatenatedString(ctx, ref)
			if err != nil {
				return err
			}
		}

		// Update the ref in the store
		if err = c.Store.SetRef(ctx, ref); err != nil {
			return err
		}

		labels := ref.DefaultLabelsValues()
		labels["pipeline_id"] = strconv.Itoa(pipeline.ID)
		labels["status"] = pipeline.Status

		// If the metric does not exist yet, start with 0 instead of 1
		// this could cause some false positives in prometheus
		// when restarting the exporter otherwise
		runCount := schemas.Metric{
			Kind:   schemas.MetricKindRunCount,
			Labels: labels,
		}

		storeGetMetric(ctx, c.Store, &runCount)

		if formerPipeline.ID != 0 && formerPipeline.ID != ref.LatestPipeline.ID {
			runCount.Value++
		}

		storeSetMetric(ctx, c.Store, runCount)

		storeSetMetric(ctx, c.Store, schemas.Metric{
			Kind:   schemas.MetricKindCoverage,
			Labels: labels,
			Value:  pipeline.Coverage,
		})

		storeSetMetric(ctx, c.Store, schemas.Metric{
			Kind:   schemas.MetricKindID,
			Labels: labels,
			Value:  float64(pipeline.ID),
		})

		emitStatusMetric(
			ctx,
			c.Store,
			schemas.MetricKindStatus,
			labels,
			statusesList[:],
			pipeline.Status,
			ref.Project.OutputSparseStatusMetrics,
		)

		storeSetMetric(ctx, c.Store, schemas.Metric{
			Kind:   schemas.MetricKindDurationSeconds,
			Labels: labels,
			Value:  pipeline.DurationSeconds,
		})

		storeSetMetric(ctx, c.Store, schemas.Metric{
			Kind:   schemas.MetricKindQueuedDurationSeconds,
			Labels: labels,
			Value:  pipeline.QueuedDurationSeconds,
		})

		storeSetMetric(ctx, c.Store, schemas.Metric{
			Kind:   schemas.MetricKindTimestamp,
			Labels: labels,
			Value:  pipeline.Timestamp,
		})

		if ref.Project.Pull.Pipeline.Jobs.Enabled {
			if err := c.PullRefPipelineJobsMetrics(ctx, ref); err != nil {
				return err
			}
		}

		return nil
	}

	if ref.Project.Pull.Pipeline.Jobs.Enabled {
		if err := c.PullRefMostRecentJobsMetrics(ctx, ref); err != nil {
			return err
		}
	}

	return nil
}
