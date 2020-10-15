// +build !race

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
	config.Pull.ProjectRefsFromProjects.OnInit = true
	config.Pull.ProjectRefsMetrics.OnInit = true
	config.GarbageCollect.Projects.OnInit = true
	config.GarbageCollect.ProjectsRefs.OnInit = true
	config.GarbageCollect.ProjectsRefsMetrics.OnInit = true

	schedulerInit(context.Background())
	// TODO: Assert if it worked as expected
}
