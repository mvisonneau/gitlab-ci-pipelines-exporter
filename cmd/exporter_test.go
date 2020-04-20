package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
)

func TestRunWrongLogLevel(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "foo", "")
	set.String("log-format", "json", "")
	fmt.Println(Run(cli.NewContext(nil, set, nil)))
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "not a valid logrus Level"))
}

func TestRunWrongLogType(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "foo", "")
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "Invalid log format"))
}

func TestRunInvalidConfigFile(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "json", "")
	set.String("config", "path_does_not_exist", "")
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "couldn't open config file :"))
}

// introduce a test to check the /metrics endpoint body
func TestMetricsRegistryContainsMetricsWhenSet(t *testing.T) {
	// a custom additional metric added to the registry
	some := "test_something"
	aCounter := prometheus.NewCounter(prometheus.CounterOpts{Name: some})
	registry := newMetricsRegistry(nil, aCounter)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	metricsHandlerFor(registry, false).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Contains(t, w.Body.String(), some)
}

func TestAMetricCanBeAddedLabelDynamically(t *testing.T) {
	// a custom additional metric added to the registry
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_something"}, []string{"first", "second"})
	prometheus.MustRegister(counter)

	curriedCounter, err := counter.CurryWith(prometheus.Labels{"first": "0", "second": "something"})
	if assert.Nil(t, err) {
		assert.Contains(t, curriedCounter.WithLabelValues().Desc().String(), "something")
	}
}

func TestAVaribleConstMetricIsUpdated(t *testing.T) {
	someVars := []gitlab.PipelineVariable{{Key: "test", Value: "testval"}, {Key: "test-2", Value: "aaaa", VariableType: "env_var"}}

	counter := variableLabelledCounter(someVars)
	assert.Contains(t, counter.Desc().String(), "test")
	assert.Contains(t, counter.Desc().String(), "test-2")
	assert.NotContains(t, counter.Desc().String(), "testval")

}
