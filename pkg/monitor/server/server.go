package server

import (
	"context"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	pb "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/protobuf"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

// Server ..
type Server struct {
	pb.UnimplementedMonitorServer

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

// Serve ..
func (s *Server) Serve() {
	if s.cfg.Global.InternalMonitoringListenerAddress == nil {
		log.Info("internal monitoring listener address not set")

		return
	}

	log.WithFields(log.Fields{
		"scheme": s.cfg.Global.InternalMonitoringListenerAddress.Scheme,
		"host":   s.cfg.Global.InternalMonitoringListenerAddress.Host,
		"path":   s.cfg.Global.InternalMonitoringListenerAddress.Path,
	}).Info("internal monitoring listener set")

	grpcServer := grpc.NewServer()
	pb.RegisterMonitorServer(grpcServer, s)

	var (
		l   net.Listener
		err error
	)

	switch s.cfg.Global.InternalMonitoringListenerAddress.Scheme {
	case "unix":
		unixAddr, err := net.ResolveUnixAddr("unix", s.cfg.Global.InternalMonitoringListenerAddress.Path)
		if err != nil {
			log.WithError(err).Fatal()
		}

		if _, err := os.Stat(s.cfg.Global.InternalMonitoringListenerAddress.Path); err == nil {
			if err := os.Remove(s.cfg.Global.InternalMonitoringListenerAddress.Path); err != nil {
				log.WithError(err).Fatal()
			}
		}

		defer func(path string) {
			if err := os.Remove(path); err != nil {
				log.WithError(err).Fatal()
			}
		}(s.cfg.Global.InternalMonitoringListenerAddress.Path)

		if l, err = net.ListenUnix("unix", unixAddr); err != nil {
			log.WithError(err).Fatal()
		}

	default:
		if l, err = net.Listen(s.cfg.Global.InternalMonitoringListenerAddress.Scheme, s.cfg.Global.InternalMonitoringListenerAddress.Host); err != nil {
			log.WithError(err).Fatal()
		}
	}

	defer l.Close()

	if err = grpcServer.Serve(l); err != nil {
		log.WithError(err).Fatal()
	}
}

// GetConfig ..
func (s *Server) GetConfig(ctx context.Context, _ *pb.Empty) (*pb.Config, error) {
	return &pb.Config{
		Content: s.cfg.ToYAML(),
	}, nil
}

// GetTelemetry ..
func (s *Server) GetTelemetry(_ *pb.Empty, ts pb.Monitor_GetTelemetryServer) (err error) {
	ctx := ts.Context()
	ticker := time.NewTicker(time.Second)

	for {
		telemetry := &pb.Telemetry{
			Projects: &pb.Entity{},
			Envs:     &pb.Entity{},
			Refs:     &pb.Entity{},
			Metrics:  &pb.Entity{},
		}

		telemetry.GitlabApiUsage = float64(s.gitlabClient.RateCounter.Rate()) / float64(s.cfg.Gitlab.MaximumRequestsPerSecond)
		if telemetry.GitlabApiUsage > 1 {
			telemetry.GitlabApiUsage = 1
		}

		telemetry.GitlabApiRequestsCount = s.gitlabClient.RequestsCounter.Load()

		telemetry.GitlabApiRateLimit = float64(s.gitlabClient.RequestsRemaining) / float64(s.gitlabClient.RequestsLimit)
		if telemetry.GitlabApiRateLimit > 1 {
			telemetry.GitlabApiRateLimit = 1
		}

		telemetry.GitlabApiLimitRemaining = uint64(s.gitlabClient.RequestsRemaining)

		var queuedTasks uint64

		queuedTasks, err = s.store.CurrentlyQueuedTasksCount(ctx)
		if err != nil {
			return
		}

		telemetry.TasksBufferUsage = float64(queuedTasks) / 1000

		telemetry.TasksExecutedCount, err = s.store.ExecutedTasksCount(ctx)
		if err != nil {
			return
		}

		telemetry.Projects.Count, err = s.store.ProjectsCount(ctx)
		if err != nil {
			return
		}

		telemetry.Envs.Count, err = s.store.EnvironmentsCount(ctx)
		if err != nil {
			return
		}

		telemetry.Refs.Count, err = s.store.RefsCount(ctx)
		if err != nil {
			return
		}

		telemetry.Metrics.Count, err = s.store.MetricsCount(ctx)
		if err != nil {
			return
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards]; ok {
			telemetry.Projects.LastPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards].Last)
			telemetry.Projects.NextPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullProjectsFromWildcards].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects]; ok {
			telemetry.Projects.LastGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects].Last)
			telemetry.Projects.NextGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectProjects].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects]; ok {
			telemetry.Envs.LastPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects].Last)
			telemetry.Envs.NextPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullEnvironmentsFromProjects].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments]; ok {
			telemetry.Envs.LastGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments].Last)
			telemetry.Envs.NextGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectEnvironments].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects]; ok {
			telemetry.Refs.LastPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects].Last)
			telemetry.Refs.NextPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullRefsFromProjects].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs]; ok {
			telemetry.Refs.LastGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs].Last)
			telemetry.Refs.NextGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectRefs].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypePullMetrics]; ok {
			telemetry.Metrics.LastPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullMetrics].Last)
			telemetry.Metrics.NextPull = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypePullMetrics].Next)
		}

		if _, ok := s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics]; ok {
			telemetry.Metrics.LastGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics].Last)
			telemetry.Metrics.NextGc = timestamppb.New(s.taskSchedulingMonitoring[schemas.TaskTypeGarbageCollectMetrics].Next)
		}

		ts.Send(telemetry)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			time.Sleep(1)
		}
	}
}
