package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"

	"gopkg.in/yaml.v2"
)

type client struct {
	*gitlab.Client
	config *config
}

var (
	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"project", "ref"},
	)

	lastRunID = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		[]string{"project", "ref"},
	)

	lastRunStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		[]string{"project", "ref", "status"},
	)

	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		[]string{"project", "ref"},
	)

	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"project", "ref"},
	)
)

func init() {
	prometheus.MustRegister(lastRunDuration)
	prometheus.MustRegister(lastRunID)
	prometheus.MustRegister(lastRunStatus)
	prometheus.MustRegister(runCount)
	prometheus.MustRegister(timeSinceLastRun)
}

func configureLogging(ctx *cli.Context) error {
	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return err
	}

	return nil
}

// Run launches the exporter
func Run(ctx *cli.Context) error {
	configureLogging(ctx)

	configFile, err := ioutil.ReadFile(ctx.GlobalString("config"))
	if err != nil {
		return exit(fmt.Errorf("Couldn't open config file : %v", err.Error()), 1)
	}

	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return exit(fmt.Errorf("Unable to parse config file: %v", err.Error()), 1)
	}

	if len(cfg.Projects) < 1 && len(cfg.Wildcards) < 1 {
		return exit(fmt.Errorf("You need to configure at least one project/wildcard to poll, none given"), 1)
	}

	// Defining defaults polling intervals
	if cfg.ProjectsPollingIntervalSeconds == 0 {
		cfg.ProjectsPollingIntervalSeconds = 1800
	}

	if cfg.RefsPollingIntervalSeconds == 0 {
		cfg.RefsPollingIntervalSeconds = 300
	}

	if cfg.PipelinesPollingIntervalSeconds == 0 {
		cfg.PipelinesPollingIntervalSeconds = 30
	}

	if cfg.PipelinesMaxPollingIntervalSeconds == 0 {
		cfg.PipelinesMaxPollingIntervalSeconds = 3600
	}

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", cfg.Gitlab.URL)
	log.Infof("Polling projects every %vs", cfg.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %vs", cfg.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %vs", cfg.PipelinesPollingIntervalSeconds)

	// Configure GitLab client
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.SkipTLSVerify},
	}

	c := &client{
		gitlab.NewClient(&http.Client{Transport: httpTransport}, cfg.Gitlab.Token),
		&cfg,
	}

	c.SetBaseURL(cfg.Gitlab.URL)
	go c.pollProjects()

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(cfg.Gitlab.URL+"/users/sign_in", 5*time.Second))

	// Expose the registered metrics via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(ctx.GlobalString("listen-address"), mux))

	return exit(nil, 0)
}

func exit(err error, exitCode int) *cli.ExitError {
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError("", exitCode)
	}

	return cli.NewExitError("", 0)
}
