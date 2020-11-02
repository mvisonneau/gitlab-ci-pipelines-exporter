package exporter

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

func metricLogFields(m schemas.Metric) log.Fields {
	return log.Fields{
		"metric-kind":   m.Kind,
		"metric-labels": m.Labels,
	}
}

func storeGetMetric(m *schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.GetMetric(m); err != nil {
		log.WithFields(
			metricLogFields(*m),
		).WithField(
			"error", err.Error(),
		).Errorf("reading metric from the store")
	}
}

func storeSetMetric(m schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.SetMetric(m); err != nil {
		log.WithFields(
			metricLogFields(m),
		).WithField(
			"error", err.Error(),
		).Errorf("writing metric in the store")
	}
}

func storeDelMetric(m schemas.Metric) {
	cfgUpdateLock.RLock()
	defer cfgUpdateLock.RUnlock()

	if err := store.DelMetric(m.Key()); err != nil {
		log.WithFields(
			metricLogFields(m),
		).WithField(
			"error", err.Error(),
		).Errorf("deleting metric from the store")
	}
}
