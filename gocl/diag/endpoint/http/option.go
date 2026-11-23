package http

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Host(v string) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return err
		}

		i.listenHost = v

		return
	}
}

func (o OptionFn) Port(v int) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.listenPort = v

		return
	}
}

func (o OptionFn) Log(v logext.Logger) OptionFn {
	return fopt.WithLog(o, v)
}

func (o OptionFn) ReadinessInspection(v readiness.Collector) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.readiness = v

		return nil
	}
}

func (o OptionFn) Conf(v endpoint.Conf) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		applyConf(i, v)

		return
	}
}

func applyConf(i *impl, cfg endpoint.Conf) {
	i.listenHost = cfg.ListenHost
	i.listenPort = cfg.ListenPort

	i.livenessPathPrefix = cfg.LivenessPathPrefix
	i.metricsPathPrefix = cfg.MetricsPathPrefix
	i.metricsPusher = cfg.MetricsPusher
	i.pprofPathPrefix = cfg.PProfPathPrefix
	i.readinessPathPrefix = cfg.ReadinessPathPrefix

	if cfg.Disabled {
		i.facilities = facilityNone

		return
	}

	switch {
	case !cfg.MetricsDisabled:
		if cfg.MetricsPush {
			i.facilities |= facilityMetricsPush
		} else {
			i.facilities |= facilityMetricsListen
		}
	}

	if !cfg.ReadinessDisabled {
		i.facilities |= facilityReadiness
	}

	if !cfg.LivenessDisabled {
		i.facilities |= facilityLiveness
	}

	if !cfg.PProfDisabled {
		i.facilities |= facilityPProf
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
