package rpc

import (
	"context"
	"net"
	"net/rpc"
	"net/url"
	"os"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
	log "github.com/sirupsen/logrus"
)

// Server ..
type Server struct {
	listenAddress            *url.URL
	gitlabClient             *gitlab.Client
	cfg                      config.Config
	store                    store.Store
	taskSchedulingMonitoring map[schemas.TaskType]*monitor.TaskSchedulingStatus
}

// NewServer ..
func NewServer(
	gitlabClient *gitlab.Client,
	c config.Config,
	st store.Store,
	tsm map[schemas.TaskType]*monitor.TaskSchedulingStatus,
) (s *Server) {
	s = &Server{
		gitlabClient:             gitlabClient,
		cfg:                      c,
		store:                    st,
		taskSchedulingMonitoring: tsm,
	}

	return
}

// ServeUNIX ..
func ServeUNIX(r *Server) {
	if r.cfg.Global.InternalMonitoringListenerAddress == nil ||
		r.cfg.Global.InternalMonitoringListenerAddress.Scheme == "" ||
		r.cfg.Global.InternalMonitoringListenerAddress.Host == "" {
		log.Info("internal monitoring listener address not set")

		return
	}

	log.WithFields(log.Fields{
		"scheme": r.cfg.Global.InternalMonitoringListenerAddress.Scheme,
		"host":   r.cfg.Global.InternalMonitoringListenerAddress.Host,
	}).Info("internal monitoring listener set")

	s := rpc.NewServer()
	if err := s.Register(r); err != nil {
		log.WithError(err).Fatal()
	}

	if r.cfg.Global.InternalMonitoringListenerAddress.Scheme == "unix" {
		if _, err := os.Stat(r.cfg.Global.InternalMonitoringListenerAddress.Host); err == nil {
			if err := os.Remove(r.cfg.Global.InternalMonitoringListenerAddress.Host); err != nil {
				log.WithError(err).Fatal()
			}
		}

		defer func(path string) {
			if err := os.Remove(path); err != nil {
				log.WithError(err).Fatal()
			}
		}(r.cfg.Global.InternalMonitoringListenerAddress.Host)
	}

	l, err := net.Listen(r.cfg.Global.InternalMonitoringListenerAddress.Scheme, r.cfg.Global.InternalMonitoringListenerAddress.Host)
	if err != nil {
		log.WithError(err).Fatal()
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.WithError(err).Fatal()
		}

		go s.ServeConn(conn)
	}
}

// Config ..
func (r *Server) Config(_ string, reply *string) error {
	*reply = r.cfg.ToYAML()

	return nil
}

// Status ..
func (r *Server) Status(_ string, reply *monitor.Status) (err error) {
	ctx := context.Background()
	s := monitor.Status{}

	s.GitLabAPIUsage = float64(r.gitlabClient.RateCounter.Rate()) / float64(r.cfg.Gitlab.MaximumRequestsPerSecond)
	if s.GitLabAPIUsage > 1 {
		s.GitLabAPIUsage = 1
	}

	s.GitLabAPIRequestsCount = r.gitlabClient.RequestsCounter

	s.GitLabAPIRateLimit = float64(r.gitlabClient.RequestsRemaining) / float64(r.gitlabClient.RequestsLimit)
	if s.GitLabAPIRateLimit > 1 {
		s.GitLabAPIRateLimit = 1
	}

	s.GitLabAPILimitRemaining = r.gitlabClient.RequestsRemaining

	var queuedTasks uint64

	queuedTasks, err = r.store.CurrentlyQueuedTasksCount(ctx)
	if err != nil {
		return
	}

	s.TasksBufferUsage = float64(queuedTasks) / 1000

	s.TasksExecutedCount, err = r.store.ExecutedTasksCount(ctx)
	if err != nil {
		return
	}

	s.Projects.Count, err = r.store.ProjectsCount(ctx)
	if err != nil {
		return
	}

	s.Envs.Count, err = r.store.EnvironmentsCount(ctx)
	if err != nil {
		return
	}

	s.Refs.Count, err = r.store.RefsCount(ctx)
	if err != nil {
		return
	}

	s.Metrics.Count, err = r.store.MetricsCount(ctx)
	if err != nil {
		return
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards]; ok {
		s.Projects.LastPull = r.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards].Last
		s.Projects.NextPull = r.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects]; ok {
		s.Projects.LastGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects].Last
		s.Projects.NextGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects]; ok {
		s.Envs.LastPull = r.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects].Last
		s.Envs.NextPull = r.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments]; ok {
		s.Envs.LastGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments].Last
		s.Envs.NextGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects]; ok {
		s.Refs.LastPull = r.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects].Last
		s.Refs.NextPull = r.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs]; ok {
		s.Refs.LastGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs].Last
		s.Refs.NextGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypePullMetrics]; ok {
		s.Metrics.LastPull = r.taskSchedulingMonitoring[schemas.TaskTypePullMetrics].Last
		s.Metrics.NextPull = r.taskSchedulingMonitoring[schemas.TaskTypePullMetrics].Next
	}

	if _, ok := r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics]; ok {
		s.Metrics.LastGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics].Last
		s.Metrics.NextGC = r.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics].Next
	}

	*reply = s

	return nil
}
