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
		discoverWildcardsProjectsEvery := time.NewTicker(time.Duration(Config.WildcardsProjectsDiscoverIntervalSeconds) * time.Second)
		discoverProjectsRefsEvery := time.NewTicker(time.Duration(Config.ProjectsRefsDiscoverIntervalSeconds) * time.Second)
		pollProjectsRefsEvery := time.NewTicker(time.Duration(Config.ProjectsRefsPollingIntervalSeconds) * time.Second)

		// init
		scheduleProjectsRefsFromPipelinesDiscovery(ctx)
		scheduleWildcardsProjectsDiscovery(ctx)
		scheduleProjectsRefsDiscovery(ctx)
		scheduleProjectsRefsMostRecentPipelinesPolling(ctx)
		scheduleProjectsRefsMostRecentJobsPolling(ctx)

		// Then, waiting for the tickers to kick in
		for {
			select {
			case <-ctx.Done():
				log.Info("stopped polling orchestration")
				return
			case <-discoverWildcardsProjectsEvery.C:
				scheduleWildcardsProjectsDiscovery(ctx)
			case <-discoverProjectsRefsEvery.C:
				scheduleProjectsRefsDiscovery(ctx)
			case <-pollProjectsRefsEvery.C:
				scheduleProjectsRefsMostRecentPipelinesPolling(ctx)
				scheduleProjectsRefsMostRecentJobsPolling(ctx)
			}
		}
	}(ctx)
}

func scheduleProjectsRefsFromPipelinesDiscovery(ctx context.Context) {
	if !Config.OnInitFetchRefsFromPipelines {
		log.WithFields(
			log.Fields{
				"init-operation": true,
			},
		).Debug("not configured to fetch refs from most recent pipelines")
		return
	}

	log.WithFields(
		log.Fields{
			"init-operation": true,
		},
	).Debug("scheduling the polling of the most recent project pipelines")

	projects, err := store.Projects()
	if err != nil {
		log.Error(err)
	}

	for _, p := range projects {
		err := pollingQueue.Add(getProjectRefsFromPipelinesTask.WithArgs(ctx, p))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func scheduleWildcardsProjectsDiscovery(ctx context.Context) {
	log.WithFields(
		log.Fields{
			"total": len(Config.Wildcards),
		},
	).Info("scheduling wildcards projects discovery")

	for _, w := range Config.Wildcards {
		err := pollingQueue.Add(getProjectsFromWildcardTask.WithArgs(ctx, w))
		if err != nil {
			log.Error(err)
		}
	}
}

func scheduleProjectsRefsDiscovery(ctx context.Context) {
	projectsCount, err := store.ProjectsCount()
	if err != nil {
		log.Error(err.Error())
	}

	log.WithFields(
		log.Fields{
			"total": projectsCount,
		},
	).Info("scheduling projects refs discovery")

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

func scheduleProjectsRefsMostRecentPipelinesPolling(ctx context.Context) {
	projectsRefsCount, err := store.ProjectsRefsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"total": projectsRefsCount,
		},
	).Info("scheduling projects refs most recent pipelines polling")

	projectRefs, err := store.ProjectsRefs()
	if err != nil {
		log.Error(err)
	}

	for _, pr := range projectRefs {
		if err = pollingQueue.Add(pollProjectRefMostRecentPipelineTask.WithArgs(ctx, pr)); err != nil {
			log.Error(err)
		}
	}
}

func scheduleProjectsRefsMostRecentJobsPolling(ctx context.Context) {
	projectsRefsCount, err := store.ProjectsRefsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"total": projectsRefsCount,
		},
	).Info("scheduling projects refs most recent jobs polling")

	projectRefs, err := store.ProjectsRefs()
	if err != nil {
		log.Error(err)
	}

	for _, pr := range projectRefs {
		if err = pollingQueue.Add(pollProjectRefMostRecentJobsTask.WithArgs(ctx, pr)); err != nil {
			log.Error(err)
		}
	}
}
