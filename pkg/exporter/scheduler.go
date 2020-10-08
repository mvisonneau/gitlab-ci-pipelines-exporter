package exporter

import (
	"context"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"

	"github.com/vmihailenco/taskq/v3"
)

var (
	getProjectsFromWildcardTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getProjectsFromWildcardTask",
		Handler: func(ctx context.Context, w schemas.Wildcard) error {
			return getProjectsFromWildcard(w)
		},
	})
	getProjectRefsFromPipelinesTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getProjectRefsFromPipelinesTask",
		Handler: func(p schemas.Project) error {
			return getProjectRefsFromPipelines(p)
		},
	})
	getRefsFromProjectTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getRefsFromProjectTask",
		Handler: func(p schemas.Project) error {
			return getRefsFromProject(p)
		},
	})
	pollProjectRefMostRecentPipelineTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pollProjectRefMostRecentPipelineTask",
		Handler: func(pr schemas.ProjectRef) error {
			return pollProjectRefMostRecentPipeline(pr)
		},
	})
	pollProjectRefMostRecentJobsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pollProjectRefMostRecentJobsTask",
		Handler: func(pr schemas.ProjectRef) error {
			return pollProjectRefMostRecentJobs(pr)
		},
	})
)

// Schedule ..
func Schedule(ctx context.Context) {
	go func(ctx context.Context) {
		pullProjectsFromWildcardsTicker := time.NewTicker(time.Duration(config.Pull.ProjectsFromWildcards.IntervalSeconds()) * time.Second)
		pullProjectRefsFromBranchesTagsMergeRequestsTicker := time.NewTicker(time.Duration(config.Pull.ProjectRefsFromBranchesTagsMergeRequests.IntervalSeconds()) * time.Second)
		pullMetricsTicker := time.NewTicker(time.Duration(config.Pull.Metrics.IntervalSeconds()) * time.Second)

		// init
		if config.Pull.ProjectsFromWildcards.OnInit() {
			schedulePullProjectsFromWildcards(ctx)
		}

		if config.Pull.ProjectRefsFromBranchesTagsMergeRequests.OnInit() {
			schedulePullProjectRefsFromBranchesTagsMergeRequests(ctx)
		}

		if config.Pull.Metrics.OnInit() {
			schedulePullMetrics(ctx)
		}

		// Scheduler configuration
		if !config.Pull.ProjectsFromWildcards.Scheduled() {
			pullProjectsFromWildcardsTicker.Stop()
		}

		if !config.Pull.ProjectRefsFromBranchesTagsMergeRequests.Scheduled() {
			pullProjectRefsFromBranchesTagsMergeRequestsTicker.Stop()
		}

		if !config.Pull.Metrics.Scheduled() {
			pullMetricsTicker.Stop()
		}

		// Waiting for the tickers to kick in
		for {
			select {
			case <-ctx.Done():
				log.Info("stopped gitlab api pull orchestration")
				return
			case <-pullProjectsFromWildcardsTicker.C:
				schedulePullProjectsFromWildcards(ctx)
			case <-pullProjectRefsFromBranchesTagsMergeRequestsTicker.C:
				schedulePullProjectRefsFromBranchesTagsMergeRequests(ctx)
			case <-pullMetricsTicker.C:
				schedulePullMetrics(ctx)
			}
		}
	}(ctx)
}

func schedulePullProjectsFromWildcards(ctx context.Context) {
	log.WithFields(
		log.Fields{
			"total": len(config.Wildcards),
		},
	).Info("scheduling projects from wildcards pull")

	for _, w := range config.Wildcards {
		err := pollingQueue.Add(getProjectsFromWildcardTask.WithArgs(ctx, w))
		if err != nil {
			log.Error(err)
		}
	}
}

func schedulePullProjectRefsFromBranchesTagsMergeRequests(ctx context.Context) {
	projectsCount, err := store.ProjectsCount()
	if err != nil {
		log.Error(err.Error())
	}

	log.WithFields(
		log.Fields{
			"total": projectsCount,
		},
	).Info("scheduling projects refs from branches, tags and merge requests pull")

	projects, err := store.Projects()
	if err != nil {
		log.Error(err)
	}

	for _, p := range projects {
		if err = pollingQueue.Add(getRefsFromProjectTask.WithArgs(ctx, p)); err != nil {
			log.Error(err)
		}
	}
}

func schedulePullMetrics(ctx context.Context) {
	projectsRefsCount, err := store.ProjectsRefsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"total": projectsRefsCount,
		},
	).Info("scheduling metrics pull")

	projectRefs, err := store.ProjectsRefs()
	if err != nil {
		log.Error(err)
	}

	for _, pr := range projectRefs {
		if err = pollingQueue.Add(pollProjectRefMostRecentPipelineTask.WithArgs(ctx, pr)); err != nil {
			log.Error(err)
		}

		if pr.Pull.Pipeline.Jobs.Enabled() {
			if err = pollingQueue.Add(pollProjectRefMostRecentJobsTask.WithArgs(ctx, pr)); err != nil {
				log.Error(err)
			}
		}
	}
}
