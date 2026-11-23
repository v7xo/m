package prommetric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defaultRegistry *prometheus.Registry = prometheus.NewRegistry()

func DefaultRegistry() *prometheus.Registry {
	return defaultRegistry
}

func DefaultRegistrer() prometheus.Registerer {
	return DefaultRegistry()
}
