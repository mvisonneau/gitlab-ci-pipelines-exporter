package controller

import (
	"context"
	"testing"
)

func TestSchedulerInit(_ *testing.T) {
	resetGlobalValues()

	configureStore()
	configurePullingQueue()
	cfg.Pull.ProjectsFromWildcards.OnInit = true
	cfg.Pull.EnvironmentsFromProjects.OnInit = true
	cfg.Pull.RefsFromProjects.OnInit = true
	cfg.Pull.Metrics.OnInit = true
	cfg.GarbageCollect.Projects.OnInit = true
	cfg.GarbageCollect.Environments.OnInit = true
	cfg.GarbageCollect.Refs.OnInit = true
	cfg.GarbageCollect.Metrics.OnInit = true

	schedulerInit(context.Background())
	// TODO: Assert if it worked as expected
}
