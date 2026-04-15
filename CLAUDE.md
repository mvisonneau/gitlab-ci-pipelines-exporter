# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code)
when working with code in this repository.

## Project Overview

A Prometheus/OpenMetrics exporter that monitors GitLab CI
pipelines, jobs, and environments.
Module: `github.com/mvisonneau/gitlab-ci-pipelines-exporter`.
Go 1.25.5.

## Commands

```bash
make fmt       # Format code (golangci-lint fmt)
make lint      # Run linters (golangci-lint run)
make test      # Run tests with race detector + coverage
make install   # Build and install binary locally
make build     # Build binaries via goreleaser snapshot
make all       # lint + test + build + coverage
make protoc    # Regenerate protobuf
make dev-env   # Docker-based dev env on port 8080
```

Run a single test:

```bash
go test -v -run TestFunctionName ./pkg/controller/
```

## Architecture

### Data Flow (cascading task-based collection)

```text
Wildcards → Projects → Refs (branches/tags/MRs)
                       → Pipelines → Jobs
                     → Environments → Deployments
```

The **Controller** (`pkg/controller/`) is the central
orchestrator. It owns config, GitLab client, store, task
queue, and rate limiter. All collection is async via `taskq`
— either Redis-backed (`redisq`) for clustering or in-memory
(`memqueue`) for single-instance.

### Key Packages

- `cmd/gitlab-ci-pipelines-exporter` — Entry point,
  delegates to `internal/cli`
- `internal/cli` — CLI framework and command definitions
- `internal/cmd` — `Run()` startup: config load →
  controller init → HTTP server → scheduler
- `pkg/controller` — Core orchestrator: scheduling,
  pulling, garbage collection, metrics handler
- `pkg/config` — YAML config parsing with hierarchical
  defaults (global → project)
- `pkg/schemas` — Data models: Project, Ref, Environment,
  Pipeline, Job, Metric. Entities use CRC32 hash keys
- `pkg/store` — Dual-backend store interface: Redis
  (`redis.go`) or local in-memory (`local.go`)
- `pkg/gitlab` — GitLab API wrapper with rate limiting,
  health checks, OpenTelemetry tracing
- `pkg/ratelimit` — Rate limiting: Redis-distributed or
  local token bucket
- `pkg/monitor` — gRPC monitoring server for task health

### HTTP Endpoints (`internal/cmd/run.go`)

- `/metrics` — Prometheus metrics
- `/health/live`, `/health/ready` — health probes
  (readiness checks GitLab reachability)
- `/webhook` — GitLab webhook ingestion
- `/debug/pprof/*` — profiling

### Task Types (`pkg/schemas/tasks.go`)

14 task types covering: pull wildcards/projects, pull
refs/environments from projects, pull ref/environment
metrics, and garbage collection for each entity type. Tasks
are scheduled periodically and cascade — top-level tasks
spawn sub-tasks for discovered entities.

## Testing Conventions

- **Framework**: `github.com/stretchr/testify/assert`
- **Mocking**: Manual HTTP mocking via `net/http/httptest`
  — no external mock frameworks
- **Pattern**: Table-driven tests with `t.Run()` subtests
- **Redis tests**: `github.com/alicebob/miniredis/v2`
  for in-memory Redis
- **Test helpers**: `newMockedGitlabAPIServer()`,
  `newTestController()` in controller test files

## Linting (`.golangci.yml`)

Enabled linters: errcheck, gosec, govet, ineffassign,
staticcheck, unused. Line length limit: 140. Import
grouping: stdlib, third-party, then
`github.com/mvisonneau` prefix (enforced by gci formatter).
