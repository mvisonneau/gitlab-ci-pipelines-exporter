package config

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := Config{}

	c.Log.Level = "info"
	c.Log.Format = "text"

	c.OpenTelemetry.GRPCEndpoint = ""

	c.Server.ListenAddress = ":8080"
	c.Server.Metrics.Enabled = true

	c.Gitlab.URL = "https://gitlab.com"
	c.Gitlab.HealthURL = "https://gitlab.com/explore"
	c.Gitlab.EnableHealthCheck = true
	c.Gitlab.EnableTLSVerify = true
	c.Gitlab.MaximumRequestsPerSecond = 1
	c.Gitlab.BurstableRequestsPerSecond = 5
	c.Gitlab.MaximumJobsQueueSize = 1000

	c.Pull.ProjectsFromWildcards.OnInit = true
	c.Pull.ProjectsFromWildcards.Scheduled = true
	c.Pull.ProjectsFromWildcards.IntervalSeconds = 1800

	c.Pull.EnvironmentsFromProjects.OnInit = true
	c.Pull.EnvironmentsFromProjects.Scheduled = true
	c.Pull.EnvironmentsFromProjects.IntervalSeconds = 1800

	c.Pull.RefsFromProjects.OnInit = true
	c.Pull.RefsFromProjects.Scheduled = true
	c.Pull.RefsFromProjects.IntervalSeconds = 300

	c.Pull.Metrics.OnInit = true
	c.Pull.Metrics.Scheduled = true
	c.Pull.Metrics.IntervalSeconds = 30

	c.GarbageCollect.Projects.Scheduled = true
	c.GarbageCollect.Projects.IntervalSeconds = 14400

	c.GarbageCollect.Environments.Scheduled = true
	c.GarbageCollect.Environments.IntervalSeconds = 14400

	c.GarbageCollect.Refs.Scheduled = true
	c.GarbageCollect.Refs.IntervalSeconds = 1800

	c.GarbageCollect.Metrics.Scheduled = true
	c.GarbageCollect.Metrics.IntervalSeconds = 600

	c.ProjectDefaults.OutputSparseStatusMetrics = true

	c.ProjectDefaults.Pull.Environments.Regexp = `.*`
	c.ProjectDefaults.Pull.Environments.ExcludeStopped = true

	c.ProjectDefaults.Pull.Refs.Branches.Enabled = true
	c.ProjectDefaults.Pull.Refs.Branches.Regexp = `^(?:main|master)$`
	c.ProjectDefaults.Pull.Refs.Branches.ExcludeDeleted = true

	c.ProjectDefaults.Pull.Refs.Tags.Enabled = true
	c.ProjectDefaults.Pull.Refs.Tags.Regexp = `.*`
	c.ProjectDefaults.Pull.Refs.Tags.ExcludeDeleted = true

	c.ProjectDefaults.Pull.Pipeline.Jobs.FromChildPipelines.Enabled = true
	c.ProjectDefaults.Pull.Pipeline.Jobs.RunnerDescription.Enabled = true
	c.ProjectDefaults.Pull.Pipeline.Jobs.RunnerDescription.AggregationRegexp = `shared-runners-manager-(\d*)\.gitlab\.com`
	c.ProjectDefaults.Pull.Pipeline.Variables.Regexp = `.*`

	assert.Equal(t, c, New())
}

func TestValidConfig(t *testing.T) {
	cfg := New()

	cfg.Gitlab.Token = "foo"
	cfg.Projects = append(cfg.Projects, NewProject("bar"))

	assert.NoError(t, cfg.Validate())
}

func TestSchedulerConfigLog(t *testing.T) {
	sc := SchedulerConfig{
		OnInit:          true,
		Scheduled:       true,
		IntervalSeconds: 300,
	}

	assert.Equal(t, log.Fields{
		"on-init":   "yes",
		"scheduled": "every 300s",
	}, sc.Log())
}
