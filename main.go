package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/heptiolabs/healthcheck"
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

	ProjectsPollingIntervalSeconds  int `yaml:"projects_polling_interval_seconds"`
	RefsPollingIntervalSeconds      int `yaml:"refs_polling_interval_seconds"`
	PipelinesPollingIntervalSeconds int `yaml:"pipelines_polling_interval_seconds"`

	DefaultRefsRegexp string `yaml:"default_refs"`

	Projects  []project
	Wildcards []wildcard
}

type client struct {
	*gitlab.Client
	config *config
}

type project struct {
	Name string
	Refs string
}

type wildcard struct {
	Search string
	Owner  struct {
		Name string
		Kind string
	}
	Refs string
}

var (
	listenAddress = flag.String("listen-address", ":8080", "Listening address")
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

func (c *client) getProject(name string) *gitlab.Project {
	p, _, err := c.Projects.GetProject(name, &gitlab.GetProjectOptions{})
	if err != nil {
		log.Fatalf("Unable to fetch project '%v' from the GitLab API : %v", name, err.Error())
		os.Exit(1)
	}
	return p
}

func (c *client) pollProjectsFromWildcards() {
	for _, w := range c.config.Wildcards {
		for _, p := range c.listProjects(&w) {
			if !c.projectExists(p) {
				log.Printf("-> Found project : %s", p.Name)
				go c.pollProject(p)
				c.config.Projects = append(c.config.Projects, p)
			}
		}
	}
}

func (c *client) projectExists(p project) bool {
	for _, cp := range c.config.Projects {
		if p == cp {
			return true
		}
	}
	return false
}

func (c *client) listProjects(w *wildcard) (projects []project) {
	log.Printf("-> Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

	trueVal := true
	falseVal := false
	var gps []*gitlab.Project
	var err error

	switch w.Owner.Kind {
	case "user":
		gps, _, err = c.Projects.ListUserProjects(
			w.Owner.Name,
			&gitlab.ListProjectsOptions{
				Archived: &falseVal,
				Simple:   &trueVal,
				Search:   &w.Search,
			},
		)
	case "group":
		gps, _, err = c.Groups.ListGroupProjects(
			w.Owner.Name,
			&gitlab.ListGroupProjectsOptions{
				Archived: &falseVal,
				Simple:   &trueVal,
				Search:   &w.Search,
			},
		)
	default:
		log.Fatalf("Invalid owner kind '%s' must be either 'user' or 'group'", w.Owner.Kind)
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Unable to list projects with search pattern '%s' from the GitLab API : %v", w.Search, err.Error())
		os.Exit(1)
	}

	for _, gp := range gps {
		projects = append(
			projects,
			project{
				Name: gp.PathWithNamespace,
				Refs: w.Refs,
			},
		)
	}

	return
}

func (c *client) pollRefs(projectID int, refsRegexp string) (refs []*string, err error) {
	if len(refsRegexp) == 0 {
		if len(c.config.DefaultRefsRegexp) > 0 {
			refsRegexp = c.config.DefaultRefsRegexp
		} else {
			refsRegexp = "^master$"
		}
	}

	re := regexp.MustCompile(refsRegexp)

	branches, err := c.pollBranchNames(projectID)
	if err != nil {
		return
	}

	for _, branch := range branches {
		if re.MatchString(*branch) {
			refs = append(refs, branch)
		}
	}

	tags, err := c.pollTagNames(projectID)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if re.MatchString(*tag) {
			refs = append(refs, tag)
		}
	}

	return
}

func (c *client) pollBranchNames(projectID int) (names []*string, err error) {
	var branches []*gitlab.Branch
	branches, _, err = c.Branches.ListBranches(projectID, &gitlab.ListBranchesOptions{})
	for _, branch := range branches {
		names = append(names, &branch.Name)
	}

	return
}

func (c *client) pollTagNames(projectID int) (names []*string, err error) {
	var tags []*gitlab.Tag
	tags, _, err = c.Tags.ListTags(projectID, &gitlab.ListTagsOptions{})
	for _, tag := range tags {
		names = append(names, &tag.Name)
	}
	return
}

func (c *client) pollProject(p project) {
	var polledRefs []string
	for {
		gp := c.getProject(p.Name)
		log.Printf("-> Polling refs for project : %s", p.Name)

		refs, err := c.pollRefs(gp.ID, p.Refs)
		if err != nil {
			log.Printf("-> Could not fetch refs for project '%s'", p.Name)
		}

		if len(refs) == 0 {
			log.Printf("-> No refs found for for project '%s'", p.Name)
			return
		}

		for _, ref := range refs {
			if !refExists(polledRefs, *ref) {
				log.Printf("-> Found ref '%s' for project '%s'", *ref, p.Name)
				go c.pollProjectRef(gp, *ref)
				polledRefs = append(polledRefs, *ref)
			}
		}

		time.Sleep(time.Duration(c.config.RefsPollingIntervalSeconds) * time.Second)
	}
}

func refExists(refs []string, r string) bool {
	for _, ref := range refs {
		if r == ref {
			return true
		}
	}
	return false
}

func (c *client) pollProjectRef(gp *gitlab.Project, ref string) {
	log.Printf("--> Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)
	var lastPipeline *gitlab.Pipeline

	for {
		pipelines, _, _ := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
		if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
			if lastPipeline != nil {
				runCount.WithLabelValues(gp.PathWithNamespace, ref).Inc()
			}

			if len(pipelines) > 0 {
				lastPipeline, _, _ = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
				lastRunDuration.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(lastPipeline.Duration))

				for _, s := range []string{"success", "failed", "running"} {
					if s == lastPipeline.Status {
						status.WithLabelValues(gp.PathWithNamespace, ref, s).Set(1)
					} else {
						status.WithLabelValues(gp.PathWithNamespace, ref, s).Set(0)
					}
				}
			} else {
				log.Printf("Could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
			}
		}

		if lastPipeline != nil {
			timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
		}

		time.Sleep(time.Duration(c.config.PipelinesPollingIntervalSeconds) * time.Second)
	}
}

func (c *client) pollProjects() {
	log.Printf("-> %d project(s) configured", len(c.config.Projects))
	for _, p := range c.config.Projects {
		go c.pollProject(p)
	}

	for {
		c.pollProjectsFromWildcards()
		time.Sleep(time.Duration(c.config.ProjectsPollingIntervalSeconds) * time.Second)
	}
}

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

	if len(config.Projects) < 1 && len(config.Wildcards) < 1 {
		log.Fatalf("You need to configure at least one project/wildcard to poll, none given")
		os.Exit(1)
	}

	// Defining defaults polling intervals
	if config.ProjectsPollingIntervalSeconds == 0 {
		config.ProjectsPollingIntervalSeconds = 1800
	}

	if config.RefsPollingIntervalSeconds == 0 {
		config.RefsPollingIntervalSeconds = 300
	}

	if config.PipelinesPollingIntervalSeconds == 0 {
		config.PipelinesPollingIntervalSeconds = 30
	}

	log.Printf("-> Starting exporter")
	log.Printf("-> Configured GitLab endpoint : %s", config.Gitlab.URL)
	log.Printf("-> Polling projects every %vs", config.ProjectsPollingIntervalSeconds)
	log.Printf("-> Polling refs every %vs", config.RefsPollingIntervalSeconds)
	log.Printf("-> Polling pipelines every %vs", config.PipelinesPollingIntervalSeconds)

	c := &client{
		gitlab.NewClient(nil, config.Gitlab.Token),
		&config,
	}

	c.SetBaseURL(config.Gitlab.URL)
	c.pollProjects()

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(len(config.Projects)+20))
	health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(config.Gitlab.URL+"/users/sign_in", 5*time.Second))

	// Expose the registered metrics via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, mux))
}
