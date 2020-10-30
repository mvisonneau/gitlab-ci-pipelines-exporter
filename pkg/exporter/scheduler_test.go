package exporter

import (
	"context"
	"testing"
)

func TestSchedulerInit(_ *testing.T) {
	resetGlobalValues()

	configureStore()
	configurePullingQueue()
	config.Pull.ProjectsFromWildcards.OnInit = true
	config.Pull.EnvironmentsFromProjects.OnInit = true
	config.Pull.RefsFromProjects.OnInit = true
	config.Pull.Metrics.OnInit = true
	config.GarbageCollect.Projects.OnInit = true
	config.GarbageCollect.Environments.OnInit = true
	config.GarbageCollect.Refs.OnInit = true
	config.GarbageCollect.Metrics.OnInit = true

	schedulerInit(context.Background())
	// TODO: Assert if it worked as expected
}
