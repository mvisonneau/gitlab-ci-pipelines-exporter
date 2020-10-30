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
	pullRefsFromProjectTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pullRefsFromProjectTask",
		Handler: func(p schemas.Project) (err error) {
			// On errors, we do not want to retry these tasks
			if err := pullRefsFromProject(p); err != nil {
				log.WithFields(log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				}).Warn("pulling projects refs from project")
			}
			return
		},
	})
	pullRefsFromPipelinesTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "getRefsFromPipelinesTask",
		Handler: func(p schemas.Project) (err error) {
			// On errors, we do not want to retry these tasks
			if err := pullRefsFromPipelines(p); err != nil {
				log.WithFields(log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				}).Warn("pulling projects refs from pipelines")
			}
			return
		},
	})
	pullRefMetricsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "pullRefMetricsTask",
		Handler: func(ref schemas.Ref) (err error) {
			// On errors, we do not want to retry these tasks
			if err := pullRefMetrics(ref); err != nil {
				log.WithFields(log.Fields{
					"project-name": ref.PathWithNamespace,
					"project-ref":  ref.Ref,
					"error":        err.Error(),
				}).Warn("pulling projects refs metrics")
			}
			return
		},
	})
	garbageCollectProjectsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "garbageCollectProjectsTask",
		Handler: func() error {
			return garbageCollectProjects()
		},
	})
	garbageCollectRefsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "garbageCollectRefsTask",
		Handler: func() error {
			return garbageCollectRefs()
		},
	})
	garbageCollectMetricsTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "garbageCollectMetricsTask",
		Handler: func() error {
			return garbageCollectMetrics()
		},
	})
)

// Schedule ..
func schedule(ctx context.Context) {
	// Check if some tasks are configured to be run on start
	schedulerInit(ctx)

	go func(ctx context.Context) {
		cfgUpdateLock.RLock()
		defer cfgUpdateLock.RUnlock()

		pullProjectsFromWildcardsTicker := time.NewTicker(time.Duration(config.Pull.ProjectsFromWildcards.IntervalSeconds) * time.Second)
		pullRefsFromProjectsTicker := time.NewTicker(time.Duration(config.Pull.RefsFromProjects.IntervalSeconds) * time.Second)
		pullMetricsTicker := time.NewTicker(time.Duration(config.Pull.Metrics.IntervalSeconds) * time.Second)
		garbageCollectProjectsTicker := time.NewTicker(time.Duration(config.GarbageCollect.Projects.IntervalSeconds) * time.Second)
		garbageCollectRefsTicker := time.NewTicker(time.Duration(config.GarbageCollect.Refs.IntervalSeconds) * time.Second)
		garbageCollectMetricsTicker := time.NewTicker(time.Duration(config.GarbageCollect.Metrics.IntervalSeconds) * time.Second)

		// Ticker configuration
		if !config.Pull.ProjectsFromWildcards.Scheduled {
			pullProjectsFromWildcardsTicker.Stop()
		}

		if !config.Pull.RefsFromProjects.Scheduled {
			pullRefsFromProjectsTicker.Stop()
		}

		if !config.Pull.Metrics.Scheduled {
			pullMetricsTicker.Stop()
		}

		if !config.GarbageCollect.Projects.Scheduled {
			garbageCollectProjectsTicker.Stop()
		}

		if !config.GarbageCollect.Refs.Scheduled {
			garbageCollectRefsTicker.Stop()
		}

		if !config.GarbageCollect.Metrics.Scheduled {
			garbageCollectMetricsTicker.Stop()
		}

		// Waiting for the tickers to kick in
		for {
			select {
			case <-ctx.Done():
				log.Info("stopped gitlab api pull orchestration")
				return
			case <-pullProjectsFromWildcardsTicker.C:
				schedulePullProjectsFromWildcards(ctx)
			case <-pullRefsFromProjectsTicker.C:
				schedulePullRefsFromProjects(ctx)
			case <-pullMetricsTicker.C:
				schedulePullMetrics(ctx)
			case <-garbageCollectProjectsTicker.C:
				schedulePullMetrics(ctx)
			case <-garbageCollectRefsTicker.C:
				scheduleGarbageCollectRefs(ctx)
			case <-garbageCollectMetricsTicker.C:
				scheduleGarbageCollectMetrics(ctx)
			}
		}
	}(ctx)
}

func schedulerInit(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if config.Pull.ProjectsFromWildcards.OnInit {
		schedulePullProjectsFromWildcards(ctx)
	}

	if config.Pull.RefsFromProjects.OnInit {
		schedulePullRefsFromProjects(ctx)
	}

	if config.Pull.Metrics.OnInit {
		schedulePullMetrics(ctx)
	}

	if config.GarbageCollect.Projects.OnInit {
		scheduleGarbageCollectProjects(ctx)
	}

	if config.GarbageCollect.Refs.OnInit {
		scheduleGarbageCollectRefs(ctx)
	}

	if config.GarbageCollect.Metrics.OnInit {
		scheduleGarbageCollectMetrics(ctx)
	}
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

func schedulePullRefsFromProjects(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

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
		go schedulePullRefsFromProject(ctx, p)
	}
}

func schedulePullMetrics(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	refsCount, err := store.RefsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"project-refs-count": refsCount,
		},
	).Info("scheduling metrics pull")

	refs, err := store.Refs()
	if err != nil {
		log.Error(err)
	}

	for _, ref := range refs {
		go schedulePullRefMetrics(ctx, ref)
	}
}

func schedulePullProjectsFromWildcardTask(ctx context.Context, w schemas.Wildcard) {
	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := pullingQueue.Add(pullProjectsFromWildcardTask.WithArgs(ctx, w)); err != nil {
		log.WithFields(log.Fields{
			"wildcard-owner-kind": w.Owner.Kind,
			"wildcard-owner-name": w.Owner.Name,
			"error":               err.Error(),
		}).Error("scheduling 'projects from wildcard' pull")
	}
}

func schedulePullRefsFromPipeline(ctx context.Context, p schemas.Project) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(pullRefsFromPipelinesTask.WithArgs(ctx, p)); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project refs from pipeline' pull")
	}
}

func schedulePullRefsFromProject(ctx context.Context, p schemas.Project) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(pullRefsFromProjectTask.WithArgs(ctx, p)); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project refs from project' pull")
	}
}

func schedulePullRefMetrics(ctx context.Context, ref schemas.Ref) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(pullRefMetricsTask.WithArgs(ctx, ref)); err != nil {
		log.WithFields(log.Fields{
			"project-name": ref.Name,
			"error":        err.Error(),
		}).Error("scheduling 'project ref most recent pipeline metrics' pull")
	}
}

func scheduleGarbageCollectProjects(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(garbageCollectProjectsTask.WithArgs(ctx)); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("scheduling 'projects garbage collection' task")
	}
}

func scheduleGarbageCollectRefs(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(garbageCollectRefsTask.WithArgs(ctx)); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("scheduling 'projects refs garbage collection' task")
	}
}

func scheduleGarbageCollectMetrics(ctx context.Context) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if pullingQueue == nil {
		log.Warn("uninitialized pulling queue, cannot schedule")
		return
	}

	if err := pullingQueue.Add(garbageCollectMetricsTask.WithArgs(ctx)); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("scheduling 'metrics garbage collection' task")
	}
}
