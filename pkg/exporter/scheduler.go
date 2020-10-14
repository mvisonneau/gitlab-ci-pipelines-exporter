package exporter

import (
	"context"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"

	"github.com/vmihailenco/taskq/v3"
)

var (
	pullProjectsFromWildcardTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getProjectsFromWildcardTask",
		Handler: func(ctx context.Context, w schemas.Wildcard) error {
			return pullProjectsFromWildcard(w)
		},
	})
	pullProjectRefsFromPipelinesTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getProjectRefsFromPipelinesTask",
		Handler: func(p schemas.Project) error {
			return pullProjectRefsFromPipelines(p)
		},
	})
	pullProjectRefsFromProjectTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pullProjectRefsFromProjectTask",
		Handler: func(p schemas.Project) error {
			return pullProjectRefsFromProject(p)
		},
	})
	pullProjectRefMetricsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pullProjectRefMetricsTask",
		Handler: func(pr schemas.ProjectRef) error {
			return pullProjectRefMetrics(pr)
		},
	})
)

// Schedule ..
func schedule(ctx context.Context) {
	go func(ctx context.Context) {
		pullProjectsFromWildcardsTicker := time.NewTicker(time.Duration(config.Pull.ProjectsFromWildcards.IntervalSeconds) * time.Second)
		pullProjectRefsFromProjectsTicker := time.NewTicker(time.Duration(config.Pull.ProjectRefsFromProjects.IntervalSeconds) * time.Second)
		pullProjectRefsMetricsTicker := time.NewTicker(time.Duration(config.Pull.ProjectRefsMetrics.IntervalSeconds) * time.Second)

		// init
		if config.Pull.ProjectsFromWildcards.OnInit {
			schedulePullProjectsFromWildcards(ctx)
		}

		if config.Pull.ProjectRefsFromProjects.OnInit {
			schedulePullProjectRefsFromProjects(ctx)
		}

		if config.Pull.ProjectRefsMetrics.OnInit {
			schedulePullProjectRefsMetrics(ctx)
		}

		// Scheduler configuration
		if !config.Pull.ProjectsFromWildcards.Scheduled {
			pullProjectsFromWildcardsTicker.Stop()
		}

		if !config.Pull.ProjectRefsFromProjects.Scheduled {
			pullProjectRefsFromProjectsTicker.Stop()
		}

		if !config.Pull.ProjectRefsMetrics.Scheduled {
			pullProjectRefsMetricsTicker.Stop()
		}

		// Waiting for the tickers to kick in
		for {
			select {
			case <-ctx.Done():
				log.Info("stopped gitlab api pull orchestration")
				return
			case <-pullProjectsFromWildcardsTicker.C:
				schedulePullProjectsFromWildcards(ctx)
			case <-pullProjectRefsFromProjectsTicker.C:
				schedulePullProjectRefsFromProjects(ctx)
			case <-pullProjectRefsMetricsTicker.C:
				schedulePullProjectRefsMetrics(ctx)
			}
		}
	}(ctx)
}

func schedulePullProjectsFromWildcards(ctx context.Context) {
	log.WithFields(
		log.Fields{
			"wildcards-count": len(config.Wildcards),
		},
	).Info("scheduling projects from wildcards pull")

	for _, w := range config.Wildcards {
		go schedulePullProjectsFromWildcardTask(ctx, w)
	}
}

func schedulePullProjectRefsFromProjects(ctx context.Context) {
	projectsCount, err := store.ProjectsCount()
	if err != nil {
		log.Error(err.Error())
	}

	log.WithFields(
		log.Fields{
			"projects-count": projectsCount,
		},
	).Info("scheduling projects refs from projects pull")

	projects, err := store.Projects()
	if err != nil {
		log.Error(err)
	}

	for _, p := range projects {
		go schedulePullProjectRefsFromProject(ctx, p)
	}
}

func schedulePullProjectRefsMetrics(ctx context.Context) {
	projectsRefsCount, err := store.ProjectsRefsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"project-refs-count": projectsRefsCount,
		},
	).Info("scheduling metrics pull")

	projectRefs, err := store.ProjectsRefs()
	if err != nil {
		log.Error(err)
	}

	for _, pr := range projectRefs {
		go schedulePullProjectRefMetrics(ctx, pr)
	}
}

func schedulePullProjectsFromWildcardTask(ctx context.Context, w schemas.Wildcard) {
	if err := pullingQueue.Add(pullProjectsFromWildcardTask.WithArgs(ctx, w)); err != nil {
		log.WithFields(log.Fields{
			"wildcard-owner-kind": w.Owner.Kind,
			"wildcard-owner-name": w.Owner.Name,
			"error":               err.Error(),
		}).Error("scheduling 'projects from wildcard' pull")
	}
}

func schedulePullProjectRefsFromPipeline(ctx context.Context, p schemas.Project) {
	if err := pullingQueue.Add(pullProjectRefsFromPipelinesTask.WithArgs(ctx, p)); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project refs from pipeline' pull")
	}
}

func schedulePullProjectRefsFromProject(ctx context.Context, p schemas.Project) {
	if err := pullingQueue.Add(pullProjectRefsFromProjectTask.WithArgs(ctx, p)); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project refs from project' pull")
	}
}

func schedulePullProjectRefMetrics(ctx context.Context, pr schemas.ProjectRef) {
	if err := pullingQueue.Add(pullProjectRefMetricsTask.WithArgs(ctx, pr)); err != nil {
		log.WithFields(log.Fields{
			"project-name": pr.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project ref most recent pipeline metrics' pull")
	}
}
