package endpoint

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/prompusher"
)

func DefaultConf() Conf {
	// nolint:mnd,revive
	return Conf{
		Disabled:            false,
		ListenHost:          "0.0.0.0",
		ListenPort:          4040,
		LivenessDisabled:    false,
		LivenessPathPrefix:  "healthz",
		MetricsDisabled:     false,
		MetricsPathPrefix:   "metrics",
		MetricsPush:         false,
		MetricsPusher:       prompusher.DefaultConf(),
		PProfDisabled:       false,
		PProfPathPrefix:     "debug/pprof",
		ReadinessDisabled:   false,
		ReadinessPathPrefix: "readyz",
	}
}
