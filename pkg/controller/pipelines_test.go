package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goGitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

type pipelineCacheProbe struct {
	store.Store
	lookupIDs []int64
	writes    int
}

type pipelineWriteFailureStore struct {
	store.Store
	target      string
	afterCommit bool
	armed       bool
	writes      map[string]int
}

func (s *pipelineWriteFailureStore) write(name string, commit func() error) error {
	s.writes[name]++
	if s.armed && s.target == name {
		s.armed = false
		if s.afterCommit {
			if err := commit(); err != nil {
				return err
			}
		}
		return fmt.Errorf("injected %s write failure", name)
	}
	return commit()
}

func (s *pipelineWriteFailureStore) SetPipeline(ctx context.Context, p schemas.Pipeline) error {
	return s.write("cache", func() error { return s.Store.SetPipeline(ctx, p) })
}

func (s *pipelineWriteFailureStore) SetRef(ctx context.Context, ref schemas.Ref) error {
	return s.write("ref", func() error { return s.Store.SetRef(ctx, ref) })
}

func (s *pipelineWriteFailureStore) SetMetric(ctx context.Context, metric schemas.Metric) error {
	name := "metric"
	if metric.Kind == schemas.MetricKindID {
		name = "id"
	}
	if metric.Kind == schemas.MetricKindRunCount {
		name = "run-count"
	}
	return s.write(name, func() error { return s.Store.SetMetric(ctx, metric) })
}

func TestPipelinePartialWriteRecovery(t *testing.T) {
	for _, backend := range []string{"local", "redis"} {
		for _, target := range []string{"cache", "ref", "id", "run-count"} {
			for _, afterCommit := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/after-commit=%t", backend, target, afterCommit), func(t *testing.T) {
					ctx, c, mux, srv := newTestController(config.Config{})
					defer srv.Close()
					if backend == "redis" {
						mr := miniredis.RunT(t)
						client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
						t.Cleanup(func() { _ = client.Close() })
						c.Store = store.NewRedisStore(client)
					}
					probe := &pipelineWriteFailureStore{Store: c.Store, target: target, afterCommit: afterCommit, writes: map[string]int{}}
					c.Store = probe
					requests := 0
					poll := 0
					for _, id := range []int{1, 2} {
						mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d", id), func(w http.ResponseWriter, r *http.Request) {
							requests++
							fmt.Fprintf(w, `{"id":%d,"status":"success","duration":20,"source":"push"}`, id)
						})
						mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d/jobs", id), func(w http.ResponseWriter, r *http.Request) {
							requests++
							fmt.Fprintf(w, `[{"id":%d,"name":"build","stage":"test","status":"running","duration":%d,"runner":{}}]`, id*100, poll+1)
						})
						mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d/bridges", id), func(w http.ResponseWriter, r *http.Request) {
							requests++
							fmt.Fprint(w, `[]`)
						})
					}
					p := schemas.NewProject("foo")
					p.Pull.Pipeline.Jobs.Enabled = true
					ref := schemas.NewRef(p, schemas.RefKindBranch, "main")
					require.NoError(t, c.Store.SetRef(ctx, ref))
					for poll = 0; poll < 3; poll++ {
						requests = 0
						probe.writes = map[string]int{}
						require.NoError(t, c.Store.GetRef(ctx, &ref))
						id := int64(2)
						if poll == 0 {
							id = 1
						}
						probe.armed = poll == 1
						err := c.ProcessPipelinesMetrics(ctx, ref, &goGitlab.PipelineInfo{ID: id})
						if poll != 1 {
							require.NoError(t, err)
						}
						t.Logf("poll=%d requests=%d writes=%v error=%v", poll, requests, probe.writes, err)
					}
					require.NoError(t, c.Store.GetRef(ctx, &ref))
					assert.Equal(t, int64(2), ref.LatestPipeline.ID)
					assert.Equal(t, int64(200), ref.LatestJobs["build"].ID)
					assert.Equal(t, float64(3), ref.LatestJobs["build"].DurationSeconds)
					// Upstream loses the transition when the ref commits without its counter write.
					expectedRunCount := float64(1)
					if target == "ref" && afterCommit || target == "run-count" && !afterCommit {
						expectedRunCount = 0
					}
					for kind, expected := range map[schemas.MetricKind]float64{schemas.MetricKindID: 2, schemas.MetricKindRunCount: expectedRunCount} {
						metric := schemas.Metric{Kind: kind, Labels: ref.DefaultLabelsValues()}
						require.NoError(t, c.Store.GetMetric(ctx, &metric))
						assert.Equal(t, expected, metric.Value, "metric %v", kind)
					}
				})
			}
		}
	}
}

func (s *pipelineCacheProbe) GetPipeline(ctx context.Context, p *schemas.Pipeline) error {
	s.lookupIDs = append(s.lookupIDs, p.ID)
	return s.Store.GetPipeline(ctx, p)
}

func (s *pipelineCacheProbe) SetPipeline(ctx context.Context, p schemas.Pipeline) error {
	s.writes++
	return s.Store.SetPipeline(ctx, p)
}

func TestPipelineCacheIdentityAndJobFreshness(t *testing.T) {
	for _, backend := range []string{"local", "redis"} {
		t.Run(backend, func(t *testing.T) {
			ctx, c, mux, srv := newTestController(config.Config{})
			defer srv.Close()
			if backend == "redis" {
				mr := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = client.Close() })
				c.Store = store.NewRedisStore(client)
			}
			probe := &pipelineCacheProbe{Store: c.Store}
			c.Store = probe
			poll := 0
			counts := map[string]int{}
			mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				counts[path]++
				page := r.URL.Query().Get("page")
				switch path {
				case "/api/v4/projects/foo/pipelines/1":
					duration := 30
					if poll == 3 {
						duration = 40
					}
					fmt.Fprintf(w, `{"id":1,"status":"success","duration":%d,"source":"push","updated_at":"2026-01-01T00:00:00Z"}`, duration)
				case "/api/v4/projects/foo/pipelines/1/jobs":
					if page == "1" {
						w.Header().Set("X-Page", "1")
						w.Header().Set("X-Next-Page", "2")
						status := []string{"running", "success", "running", "success"}[poll]
						id := 10
						if poll >= 2 {
							id = 20
						}
						fmt.Fprintf(w, `[{"id":%d,"name":"build","stage":"test","status":%q,"duration":%d,"runner":{},"tag_list":["linux"]}]`, id, status, poll+1)
					} else {
						w.Header().Set("X-Page", "2")
						status := "manual"
						if poll > 0 {
							status = "running"
						}
						fmt.Fprintf(w, `[{"id":11,"name":"manual","stage":"deploy","status":%q,"runner":{}}`, status)
						if poll > 0 {
							fmt.Fprint(w, `,{"id":12,"name":"new-job","stage":"test","status":"success","runner":{}}`)
						}
						fmt.Fprint(w, `,{"id":13,"name":"same-name","stage":"first","status":"success","runner":{}},{"id":14,"name":"same-name","stage":"second","status":"success","runner":{}}]`)
					}
				case "/api/v4/projects/foo/pipelines/1/bridges":
					if page == "1" {
						w.Header().Set("X-Page", "1")
						w.Header().Set("X-Next-Page", "2")
						fmt.Fprint(w, `[{"id":90}]`)
					} else {
						w.Header().Set("X-Page", "2")
						if poll == 0 {
							fmt.Fprint(w, `[{"id":91}]`)
						} else {
							fmt.Fprint(w, `[{"id":91,"downstream_pipeline":{"id":2,"project_id":99}}]`)
						}
					}
				case "/api/v4/projects/99/pipelines/2/jobs":
					fmt.Fprintf(w, `[{"id":30,"name":"child","stage":"test","status":%q,"duration":%d,"runner":{}}]`, []string{"pending", "running", "success", "success"}[poll], poll)
				case "/api/v4/projects/99/pipelines/2/bridges":
					fmt.Fprint(w, `[]`)
				case "/api/v4/projects/foo/jobs":
					fmt.Fprint(w, `[{"id":10,"name":"build","ref":"main","stage":"test","status":"success","duration":2,"runner":{},"tag_list":["linux"]}]`)
				default:
					t.Errorf("unexpected request: %s", r.URL)
					http.Error(w, "unexpected request", http.StatusNotFound)
				}
			})
			p := schemas.NewProject("foo")
			p.Pull.Pipeline.Jobs.Enabled = true
			ref := schemas.NewRef(p, schemas.RefKindBranch, "main")
			require.NoError(t, c.Store.SetRef(ctx, ref))
			for poll = 0; poll < 4; poll++ {
				counts = map[string]int{}
				require.NoError(t, c.Store.GetRef(ctx, &ref))
				require.NoError(t, c.ProcessPipelinesMetrics(ctx, ref, &goGitlab.PipelineInfo{ID: 1}))
				require.NoError(t, c.Store.GetRef(ctx, &ref))
				assert.Equal(t, []string{"running", "success", "running", "success"}[poll], ref.LatestJobs["build"].Status)
				assert.Equal(t, float64(poll+1), ref.LatestJobs["build"].DurationSeconds)
				assert.Equal(t, int64(14), ref.LatestJobs["same-name"].ID)
				assert.Equal(t, "success", ref.LatestPipeline.Status)
				assert.Equal(t, float64(1767225600), ref.LatestPipeline.Timestamp)
				metrics, err := c.Store.Metrics(ctx)
				require.NoError(t, err)
				labels := ref.DefaultLabelsValues()
				labels["stage"] = "test"
				labels["job_name"] = "build"
				labels["tag_list"] = "linux"
				labels["failure_reason"] = ""
				labels["runner_description"] = ""
				metric := schemas.Metric{Kind: schemas.MetricKindJobDurationSeconds, Labels: labels}
				assert.Equal(t, float64(poll+1), metrics[metric.Key()].Value)
				assert.Equal(t, labels, map[string]string(metrics[metric.Key()].Labels))
				assert.Equal(t, 2, counts["/api/v4/projects/foo/pipelines/1/jobs"])
				assert.Equal(t, 2, counts["/api/v4/projects/foo/pipelines/1/bridges"])
				assert.Zero(t, counts["/api/v4/projects/foo/jobs"])
				total := 0
				for _, n := range counts {
					total += n
				}
				if poll == 0 {
					assert.Equal(t, 5, total)
					assert.Equal(t, "manual", ref.LatestJobs["manual"].Status)
				} else {
					assert.Equal(t, 7, total)
					assert.Equal(t, "running", ref.LatestJobs["manual"].Status)
					assert.Contains(t, ref.LatestJobs, "new-job")
					assert.Equal(t, float64(poll), ref.LatestJobs["child"].DurationSeconds)
					assert.Equal(t, []string{"", "running", "success", "success"}[poll], ref.LatestJobs["child"].Status)
				}
				if poll >= 2 {
					assert.Equal(t, int64(20), ref.LatestJobs["build"].ID)
				}
				t.Logf("poll=%d total GitLab requests=%d", poll, total)
			}
			assert.Equal(t, []int64{1, 1, 1, 1}, probe.lookupIDs)
			assert.Equal(t, 2, probe.writes)
			assert.Equal(t, float64(40), ref.LatestPipeline.DurationSeconds)
		})
	}
}

func TestCachedPipelineForNewRef(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()
	mux.HandleFunc("/api/v4/projects/foo/pipelines/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":1,"status":"success","source":"push"}`)
	})
	for _, name := range []string{"main", "second-ref"} {
		ref := schemas.NewRef(schemas.NewProject("foo"), schemas.RefKindBranch, name)
		require.NoError(t, c.Store.SetRef(ctx, ref))
		require.NoError(t, c.ProcessPipelinesMetrics(ctx, ref, &goGitlab.PipelineInfo{ID: 1}))
		require.NoError(t, c.Store.GetRef(ctx, &ref))
		assert.Equal(t, int64(1), ref.LatestPipeline.ID)
		metric := schemas.Metric{Kind: schemas.MetricKindID, Labels: ref.DefaultLabelsValues()}
		require.NoError(t, c.Store.GetMetric(ctx, &metric))
		assert.Equal(t, float64(1), metric.Value)
	}
}

func TestCachedMultiplePipelinesKeepNewestRef(t *testing.T) {
	for _, backend := range []string{"local", "redis"} {
		t.Run(backend, func(t *testing.T) {
			ctx, c, mux, srv := newTestController(config.Config{})
			defer srv.Close()
			if backend == "redis" {
				mr := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = client.Close() })
				c.Store = store.NewRedisStore(client)
			}
			poll := 0
			requests := 0
			mux.HandleFunc("/api/v4/projects/foo/pipelines", func(w http.ResponseWriter, r *http.Request) {
				requests++
				assert.Equal(t, "2", r.URL.Query().Get("per_page"))
				assert.Equal(t, "main", r.URL.Query().Get("ref"))
				fmt.Fprint(w, `[{"id":2},{"id":1}]`)
			})
			for _, id := range []int{1, 2} {
				mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d", id), func(w http.ResponseWriter, r *http.Request) {
					requests++
					status := "failed"
					if id == 2 {
						status = "success"
					}
					fmt.Fprintf(w, `{"id":%d,"status":%q,"duration":%d,"source":"push","updated_at":"2026-01-0%dT00:00:00Z"}`, id, status, id*10, id)
				})
				mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d/jobs", id), func(w http.ResponseWriter, r *http.Request) {
					requests++
					fmt.Fprintf(w, `[{"id":%d,"name":"build","stage":"test","status":"running","duration":%d,"runner":{}}]`, id*10, poll+1)
				})
				mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/%d/bridges", id), func(w http.ResponseWriter, r *http.Request) {
					requests++
					fmt.Fprint(w, `[]`)
				})
			}
			p := schemas.NewProject("foo")
			p.Pull.Pipeline.PerRef = 2
			p.Pull.Pipeline.Jobs.Enabled = true
			ref := schemas.NewRef(p, schemas.RefKindBranch, "main")
			require.NoError(t, c.Store.SetRef(ctx, ref))
			for poll = 0; poll < 3; poll++ {
				requests = 0
				require.NoError(t, c.PullRefMetrics(ctx, ref))
				require.NoError(t, c.Store.GetRef(ctx, &ref))
				assert.Equal(t, int64(2), ref.LatestPipeline.ID, "poll %d", poll)
				assert.Equal(t, int64(20), ref.LatestJobs["build"].ID)
				assert.Equal(t, float64(poll+1), ref.LatestJobs["build"].DurationSeconds)
				assert.Equal(t, 7, requests)
				for kind, expected := range map[schemas.MetricKind]float64{
					schemas.MetricKindID:              2,
					schemas.MetricKindDurationSeconds: 20,
					schemas.MetricKindTimestamp:       1767312000,
					schemas.MetricKindRunCount:        float64(poll),
				} {
					metric := schemas.Metric{Kind: kind, Labels: ref.DefaultLabelsValues()}
					require.NoError(t, c.Store.GetMetric(ctx, &metric))
					assert.Equal(t, expected, metric.Value, "poll %d, metric %v", poll, kind)
				}
				labels := ref.DefaultLabelsValues()
				labels["status"] = "success"
				metric := schemas.Metric{Kind: schemas.MetricKindStatus, Labels: labels}
				require.NoError(t, c.Store.GetMetric(ctx, &metric))
				assert.Equal(t, float64(1), metric.Value)
			}
		})
	}
}

func TestCachedMergeRequestPipelineFallback(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()
	var refs []string
	details := 0
	mux.HandleFunc("/api/v4/projects/foo/pipelines", func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("ref")
		refs = append(refs, ref)
		assert.Equal(t, "1", r.URL.Query().Get("per_page"))
		if ref == "refs/merge-requests/1234/head" {
			fmt.Fprint(w, `[]`)
		} else {
			assert.Equal(t, "refs/merge-requests/1234/merge", ref)
			fmt.Fprint(w, `[{"id":1}]`)
		}
	})
	mux.HandleFunc("/api/v4/projects/foo/pipelines/1", func(w http.ResponseWriter, r *http.Request) {
		details++
		fmt.Fprint(w, `{"id":1,"status":"success","source":"merge_request_event"}`)
	})
	ref := schemas.NewRef(schemas.NewProject("foo"), schemas.RefKindMergeRequest, "1234")
	require.NoError(t, c.Store.SetRef(ctx, ref))
	for range 2 {
		require.NoError(t, c.PullRefMetrics(ctx, ref))
	}
	assert.Equal(t, []string{"refs/merge-requests/1234/head", "refs/merge-requests/1234/merge", "refs/merge-requests/1234/head", "refs/merge-requests/1234/merge"}, refs)
	assert.Equal(t, 2, details)
	require.NoError(t, c.Store.GetRef(ctx, &ref))
	assert.Equal(t, int64(1), ref.LatestPipeline.ID)
}

func TestPullRefMetricsSucceed(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.URL.Query().Get("ref"))
			_, _ = fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"id":1,"created_at":"2016-08-11T11:27:00.085Z", "started_at":"2016-08-11T11:28:00.085Z",
			"duration":300,"queued_duration":60,"status":"running","coverage":"30.2","source":"schedule"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.TestReports.Enabled = true
	p.Pull.Pipeline.TestReports.TestCases.Enabled = true

	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindBranch,
			"bar",
		)))

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics(ctx)
	labels := map[string]string{
		"kind":      string(schemas.RefKindBranch),
		"project":   "foo",
		"ref":       "bar",
		"topics":    "",
		"variables": "foo:bar",
		"source":    "schedule",
	}

	runCount := schemas.Metric{
		Kind:   schemas.MetricKindRunCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, runCount, metrics[runCount.Key()])

	coverage := schemas.Metric{
		Kind:   schemas.MetricKindCoverage,
		Labels: labels,
		Value:  30.2,
	}
	assert.Equal(t, coverage, metrics[coverage.Key()])

	runID := schemas.Metric{
		Kind:   schemas.MetricKindID,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, runID, metrics[runID.Key()])

	queued := schemas.Metric{
		Kind:   schemas.MetricKindQueuedDurationSeconds,
		Labels: labels,
		Value:  60,
	}
	assert.Equal(t, queued, metrics[queued.Key()])

	labels["status"] = "running"
	status := schemas.Metric{
		Kind:   schemas.MetricKindStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])
}

func TestPullRefMetricsUpdatingPipeline(t *testing.T) {
	// given
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()
	apiPipeline := `{
		"id":1,
		"created_at":"2016-08-11T11:27:00.085Z",
		"started_at":"2016-08-11T11:28:00.085Z",
		"duration":300,
		"queued_duration":60,
		"status":"running",
		"coverage":"30.2",
		"source":"pipeline"
	}`

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.URL.Query().Get("ref"))
			_, _ = fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, apiPipeline)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true

	labels := map[string]string{
		"kind":      string(schemas.RefKindBranch),
		"project":   "foo",
		"ref":       "bar",
		"topics":    "",
		"variables": "foo:bar",
		"source":    "pipeline",
	}

	// when
	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindBranch,
			"bar",
		)))

	metrics, _ := c.Store.Metrics(ctx)

	// then
	runID := schemas.Metric{
		Kind:   schemas.MetricKindID,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, runID, metrics[runID.Key()])

	labels["status"] = "running"
	status := schemas.Metric{
		Kind:   schemas.MetricKindStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])

	// given again
	apiPipeline = `{
		"id":1,
		"created_at":"2016-08-11T11:27:00.085Z",
		"started_at":"2016-08-11T11:28:00.085Z",
		"duration":300,
		"queued_duration":60,
		"status":"failed",
		"coverage":"30.2",
		"source":"pipeline"
	}`

	labels = map[string]string{
		"kind":      string(schemas.RefKindBranch),
		"project":   "foo",
		"ref":       "bar",
		"topics":    "",
		"variables": "foo:bar",
		"source":    "pipeline",
	}

	// when again
	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindBranch,
			"bar",
		)))

	metrics, _ = c.Store.Metrics(ctx)

	// then again
	runID = schemas.Metric{
		Kind:   schemas.MetricKindID,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, runID, metrics[runID.Key()])

	labels["status"] = "failed"
	status = schemas.Metric{
		Kind:   schemas.MetricKindStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])
}

func TestPullRefTestReportMetrics(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.URL.Query().Get("ref"))
			_, _ = fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"id":1,"created_at":"2016-08-11T11:27:00.085Z", "started_at":"2016-08-11T11:28:00.085Z",
			"duration":300,"queued_duration":60,"status":"success","coverage":"30.2","source":"schedule"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.TestReports.Enabled = true
	p.Pull.Pipeline.TestReports.TestCases.Enabled = true

	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindBranch,
			"bar",
		)))

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics(ctx)
	labels := map[string]string{
		"kind":      string(schemas.RefKindBranch),
		"project":   "foo",
		"ref":       "bar",
		"topics":    "",
		"variables": "foo:bar",
		"source":    "schedule",
	}

	trTotalTime := schemas.Metric{
		Kind:   schemas.MetricKindTestReportTotalTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, trTotalTime, metrics[trTotalTime.Key()])

	trTotalCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportTotalCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, trTotalCount, metrics[trTotalCount.Key()])

	trSuccessCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportSuccessCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, trSuccessCount, metrics[trSuccessCount.Key()])

	trFailedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportFailedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trFailedCount, metrics[trFailedCount.Key()])

	trSkippedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportSkippedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trSkippedCount, metrics[trSkippedCount.Key()])

	trErrorCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportErrorCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trErrorCount, metrics[trErrorCount.Key()])

	labels["test_suite_name"] = "Secure"

	tsTotalTime := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteTotalTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, tsTotalTime, metrics[tsTotalTime.Key()])

	tsTotalCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteTotalCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tsTotalCount, metrics[tsTotalCount.Key()])

	tsSuccessCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteSuccessCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tsSuccessCount, metrics[tsSuccessCount.Key()])

	tsFailedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteFailedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsFailedCount, metrics[tsFailedCount.Key()])

	tsSkippedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteSkippedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsSkippedCount, metrics[tsSkippedCount.Key()])

	tsErrorCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteErrorCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsErrorCount, metrics[tsErrorCount.Key()])

	labels["test_case_name"] = "Security Reports can create an auto-remediation MR"
	labels["test_case_classname"] = "vulnerability_management_spec"

	tcExecutionTime := schemas.Metric{
		Kind:   schemas.MetricKindTestCaseExecutionTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, tcExecutionTime, metrics[tcExecutionTime.Key()])

	labels["status"] = "success"
	tcStatus := schemas.Metric{
		Kind:   schemas.MetricKindTestCaseStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tcStatus, metrics[tcStatus.Key()])
}

func TestPullRefMetricsMergeRequestPipeline(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "refs/merge-requests/1234/head", r.URL.Query().Get("ref"))
			_, _ = fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"id":1,"updated_at":"2016-08-11T11:28:34.085Z","duration":300,"status":"running","coverage":"30.2","source":"schedule"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			_, _ = fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true

	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindMergeRequest,
			"1234",
		)))
}
