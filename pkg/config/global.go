package config

import (
	"net/url"
)

// Global is used for globally shared exporter config.
type Global struct {
	// InternalMonitoringListenerAddress can be used to access
	// some metrics related to the exporter internals
	InternalMonitoringListenerAddress *url.URL
}
