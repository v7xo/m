package server

import (
	"net/http"

	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/cors"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logging.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Host(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.listenHost = val })
}

func (o OptionFn) PreventErrDetailedMessage(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.preventErrDetailedMessage = val })
}

func (o OptionFn) Port(val int) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.listenPort = val })
}

func (o OptionFn) Conf(val Conf) OptionFn {
	return fopt.Chain(o, func(t *impl) { applyConf(t, val) })
}

func (o OptionFn) BeforeStartServe(val DriverFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.beforeStartServe = val })
}

func (o OptionFn) CORSEnabled(val bool) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			if val {
				t.corsOpts = cors.Option()
			} else {
				t.corsOpts = nil
			}
		},
	)
}

func (o OptionFn) Handler(val http.Handler) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.handler = val })
}

func applyConf(i *impl, c Conf) {
	i.shutdownTimeout = timeout.Timeout(c.ShutdownTimeoutDur)
	i.listenHost = c.ListenHost
	i.listenPort = c.ListenPort
	i.preventErrDetailedMessage = c.PreventErrDetailedMessage

	if c.CORSEnabled {
		i.corsOpts = cors.Option().WithConfig(c.CORS)
	} else {
		i.corsOpts = nil
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
