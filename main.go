package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xanzy/go-gitlab"

	"gopkg.in/yaml.v2"
)

type config struct {
	Gitlab struct {
		URL   string
		Token string
	}

	PollingIntervalSeconds int `yaml:"polling_interval_seconds"`
	Projects               []project
}

type project struct {
	Name string
	Ref  string
}

var (
	listenAddress = flag.String("listen-address", ":80", "Listening address")
	configPath    = flag.String("config", "~/.gitlab-ci-pipelines-exporter.yml", "Config file path")
)

var (
	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"project", "ref"},
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"project", "ref"},
	)
	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		[]string{"project", "ref"},
	)

	status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_status",
			Help: "GitLab CI pipeline current status",
		},
		[]string{"project", "ref", "status"},
	)
)

func init() {
	prometheus.MustRegister(timeSinceLastRun)
	prometheus.MustRegister(lastRunDuration)
	prometheus.MustRegister(runCount)
	prometheus.MustRegister(status)
}

func main() {
	flag.Parse()

	var config config

	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Couldn't open config file : %v", err.Error())
		os.Exit(1)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err.Error())
		os.Exit(1)
	}

	if len(config.Projects) < 1 {
		log.Fatalf("You need to configure at least one project/ref to poll, none given")
		os.Exit(1)
	}

	log.Printf("-> Starting exporter")
	log.Printf("-> Polling %v every %vs", config.Gitlab.URL, config.PollingIntervalSeconds)
	log.Printf("-> %v project(s) configured", len(config.Projects))

	gc := gitlab.NewClient(nil, config.Gitlab.Token)
	gc.SetBaseURL(config.Gitlab.URL)

	for _, project := range config.Projects {
		go func(project project) {
			p, _, err := gc.Projects.GetProject(project.Name)
			if err != nil {
				log.Fatalf("Unable to fetch project '%v' from the GitLab API : %v", project.Name, err.Error())
				os.Exit(1)
			}

			log.Printf("--> Polling ID: %v | %v:%v", p.ID, project.Name, project.Ref)

			var lastPipeline *gitlab.Pipeline
			runCount.WithLabelValues(project.Name, project.Ref).Set(0)

			for {
				pipelines, _, _ := gc.Pipelines.ListProjectPipelines(p.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(project.Ref)})
				if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
					if lastPipeline != nil {
						runCount.WithLabelValues(project.Name, project.Ref).Inc()
					}

					lastPipeline, _, _ = gc.Pipelines.GetPipeline(p.ID, pipelines[0].ID)

					lastRunDuration.WithLabelValues(project.Name, project.Ref).Set(float64(lastPipeline.Duration))

					for _, s := range []string{"success", "failed", "running"} {
						if s == lastPipeline.Status {
							status.WithLabelValues(project.Name, project.Ref, s).Set(1)
						} else {
							status.WithLabelValues(project.Name, project.Ref, s).Set(0)
						}
					}
				}

				timeSinceLastRun.WithLabelValues(project.Name, project.Ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))

				time.Sleep(time.Duration(config.PollingIntervalSeconds) * time.Second)
			}
		}(project)
	}

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
