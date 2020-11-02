package exporter

import "time"

// CacheEnvironmentCommitInfo ..
type CacheEnvironmentCommitInfo struct {
	ShortID    string
	CreatedAt  time.Time
	ValidUntil time.Time
}

// CacheEnvironmentRefQuery ..
type CacheEnvironmentRefQuery struct {
	ProjectName string
	Search      string
}

// CacheEnvironmentRefs ..
type CacheEnvironmentRefs map[CacheEnvironmentRefQuery]CacheEnvironmentCommitInfo

var (
	cacheEnvironmentRefs CacheEnvironmentRefs
)
