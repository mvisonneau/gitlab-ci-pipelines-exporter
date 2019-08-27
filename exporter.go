package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"

	"gopkg.in/yaml.v2"
)

type config struct {
	Gitlab struct {
		URL           string
		Token         string
		SkipTLSVerify bool `yaml:"skip_tls_verify"`
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
				log.Infof("Found project : %s", p.Name)
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
	log.Infof("Listing all projects using search pattern : '%s' with owner '%s' (%s)", w.Search, w.Owner.Name, w.Owner.Kind)

	trueVal := true
	falseVal := false
	listOptions := gitlab.ListOptions{
		PerPage: 20,
		Page:    1,
	}

	for {
		var gps []*gitlab.Project
		var resp *gitlab.Response
		var err error

		switch w.Owner.Kind {
		case "user":
			gps, resp, err = c.Projects.ListUserProjects(
				w.Owner.Name,
				&gitlab.ListProjectsOptions{
					ListOptions: listOptions,
					Archived:    &falseVal,
					Simple:      &trueVal,
					Search:      &w.Search,
				},
			)
		case "group":
			gps, resp, err = c.Groups.ListGroupProjects(
				w.Owner.Name,
				&gitlab.ListGroupProjectsOptions{
					ListOptions: listOptions,
					Archived:    &falseVal,
					Simple:      &trueVal,
					Search:      &w.Search,
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

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		listOptions.Page = resp.NextPage
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

func (c *client) pollBranchNames(projectID int) ([]*string, error) {
	var names []*string

	options := &gitlab.ListBranchesOptions{
		PerPage: 20,
		Page:    1,
	}

	for {
		branches, resp, err := c.Branches.ListBranches(projectID, options)
		if err != nil {
			return names, err
		}

		for _, branch := range branches {
			names = append(names, &branch.Name)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *client) pollTagNames(projectID int) ([]*string, error) {
	var names []*string

	options := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	for {
		tags, resp, err := c.Tags.ListTags(projectID, options)
		if err != nil {
			return names, err
		}

		for _, tag := range tags {
			names = append(names, &tag.Name)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		options.Page = resp.NextPage
	}

	return names, nil
}

func (c *client) pollProject(p project) {
	var polledRefs []string
	for {
		gp := c.getProject(p.Name)
		log.Infof("Polling refs for project : %s", p.Name)

		refs, err := c.pollRefs(gp.ID, p.Refs)
		if err != nil {
			log.Warnf("Could not fetch refs for project '%s'", p.Name)
		}

		if len(refs) == 0 {
			log.Warnf("No refs found for for project '%s'", p.Name)
			return
		}

		for _, ref := range refs {
			if !refExists(polledRefs, *ref) {
				log.Infof("Found ref '%s' for project '%s'", *ref, p.Name)
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
	log.Infof("Polling %v:%v (%v)", gp.PathWithNamespace, ref, gp.ID)
	var lastPipeline *gitlab.Pipeline

	runCount.WithLabelValues(gp.PathWithNamespace, ref).Add(0)

	for {
		pipelines, _, _ := c.Pipelines.ListProjectPipelines(gp.ID, &gitlab.ListProjectPipelinesOptions{Ref: gitlab.String(ref)})
		if lastPipeline == nil || lastPipeline.ID != pipelines[0].ID || lastPipeline.Status != pipelines[0].Status {
			if lastPipeline != nil {
				runCount.WithLabelValues(gp.PathWithNamespace, ref).Inc()
			}

			if len(pipelines) > 0 {
				lastPipeline, _, _ = c.Pipelines.GetPipeline(gp.ID, pipelines[0].ID)
				lastRunDuration.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(lastPipeline.Duration))
				lastRunID.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(lastPipeline.ID))

				for _, s := range []string{"success", "failed", "running"} {
					if s == lastPipeline.Status {
						lastRunStatus.WithLabelValues(gp.PathWithNamespace, ref, s).Set(1)
					} else {
						lastRunStatus.WithLabelValues(gp.PathWithNamespace, ref, s).Set(0)
					}
				}
			} else {
				log.Warnf("Could not find any pipeline for %s:%s", gp.PathWithNamespace, ref)
			}
		}

		if lastPipeline != nil {
			timeSinceLastRun.WithLabelValues(gp.PathWithNamespace, ref).Set(float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
		}

		time.Sleep(time.Duration(c.config.PipelinesPollingIntervalSeconds) * time.Second)
	}
}

func (c *client) pollProjects() {
	log.Infof("%d project(s) configured", len(c.config.Projects))
	for _, p := range c.config.Projects {
		go c.pollProject(p)
	}

	for {
		c.pollProjectsFromWildcards()
		time.Sleep(time.Duration(c.config.ProjectsPollingIntervalSeconds) * time.Second)
	}
}

func init() {
	prometheus.MustRegister(lastRunDuration)
	prometheus.MustRegister(lastRunID)
	prometheus.MustRegister(lastRunStatus)
	prometheus.MustRegister(runCount)
	prometheus.MustRegister(timeSinceLastRun)
}

func run(ctx *cli.Context) error {
	configureLogging(ctx.GlobalString("log-level"), ctx.GlobalString("log-format"))

	var config config

	configFile, err := ioutil.ReadFile(ctx.GlobalString("config"))
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

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", config.Gitlab.URL)
	log.Infof("Polling projects every %vs", config.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %vs", config.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %vs", config.PipelinesPollingIntervalSeconds)

	// Configure GitLab client
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Gitlab.SkipTLSVerify},
	}

	c := &client{
		gitlab.NewClient(&http.Client{Transport: httpTransport}, config.Gitlab.Token),
		&config,
	}

	c.SetBaseURL(config.Gitlab.URL)
	go c.pollProjects()

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(config.Gitlab.URL+"/users/sign_in", 5*time.Second))

	// Expose the registered metrics via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(ctx.GlobalString("listen-address"), mux))

	return nil
}
