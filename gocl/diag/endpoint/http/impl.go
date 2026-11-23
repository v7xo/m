package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prompusher"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type facilities int64

const (
	facilityMetricsListen facilities = facilities(1 << iota)
	facilityMetricsPush
	facilityPProf
	facilityReadiness
	facilityLiveness

	facilityNone = facilities(0)
)

func (v facilities) Metrics() bool {
	return v&(facilityMetricsListen|facilityMetricsPush) != facilityNone
}

func (v facilities) MetricsPush() bool {
	return v&facilityMetricsPush == facilityMetricsPush
}

func (v facilities) MetricsListen() bool {
	return v&facilityMetricsListen == facilityMetricsListen
}

func (v facilities) Readiness() bool {
	return v&facilityReadiness == facilityReadiness
}

func (v facilities) Liveness() bool {
	return v&facilityLiveness == facilityLiveness
}

func (v facilities) PProf() bool {
	return v&facilityPProf == facilityPProf
}

func (v facilities) None() bool {
	return v == facilityNone
}

type impl struct {
	fopt.LogOpt

	readiness readiness.Collector

	listenHost string

	livenessPathPrefix  string
	metricsPathPrefix   string
	pprofPathPrefix     string
	readinessPathPrefix string

	metricsPusher prompusher.Conf

	listenPort int
	facilities facilities
}

func (s *impl) Start(stop context.Context) (err error) {
	if s.facilities.None() {
		return
	}

	var httpMux *http.ServeMux

	if httpMux, err = s.makeHTTPMux(stop); err != nil {
		return
	}

	var pusher prompusher.Pusher

	switch {
	case s.facilities.MetricsPush():
		if pusher, err = prompusher.New(prompusher.Option().Log(s.Log()).Conf(s.metricsPusher)); err != nil {
			return
		}
	}

	var (
		// TODO: assign ErrorLog
		httpSrv = &http.Server{
			// nolint:godox
			Handler: httpMux,
		}

		lnCfg  net.ListenConfig
		lnAddr = net.JoinHostPort(s.listenHost, strconv.Itoa(s.listenPort))
	)

	ln, err := lnCfg.Listen(stop, "tcp", lnAddr)
	if err != nil {
		err = fmt.Errorf("failed to bind diagnostics endpoint; addr: %s; %w", lnAddr, err)

		return
	}

	go func() {
		<-stop.Done()

		if closeErr := httpSrv.Close(); closeErr != nil {
			s.Log().
				Err(stop, closeErr).
				Msg("error while closing diagnostics endpoint HTTP-server")
		}
	}()

	go func() {
		s.Log().
			Info(stop).
			Msgf("start serving diagnostics endpoint on '%s' ...", ln.Addr())

		if serveErr := httpSrv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) && !errors.Is(serveErr, net.ErrClosed) {
			s.Log().
				Err(stop, serveErr).
				Msg("error while serving diagnostics endpoint HTTP-server")
		}
	}()

	if pusher != nil {
		err = pusher.Start(stop)
	}

	return
}

func (s *impl) makeHTTPMux(stopCtx context.Context) (mux *http.ServeMux, err error) {
	mux = http.NewServeMux()

	if s.facilities.MetricsPush() && s.facilities.MetricsListen() {
		err = errors.New("it is unable to service metrics pusher and metrics listen endpoint simultaneously")

		return
	} else if s.facilities.MetricsListen() {
		var handler http.Handler

		if handler, err = makeMetricsHandler(stopCtx, s.Log()); err != nil {
			return
		}

		mux.Handle("/"+s.metricsPathPrefix, handler)
	}

	if s.facilities.Readiness() {
		mux.HandleFunc(
			"/"+s.readinessPathPrefix,
			func(w http.ResponseWriter, r *http.Request) {
				const notReadyMsg = "NOT READY"

				var (
					ctx      = r.Context()
					writeErr error
				)

				defer func() {
					if writeErr != nil {
						s.Log().
							Err(ctx, writeErr).
							Msgf("failed to write response for readiness probe request; request URI: %s", r.URL.Redacted())
					}
				}()

				readiness, readinessErr := s.readiness.CollectReadinessConsolidated(ctx)
				if readinessErr != nil {
					s.Log().
						Err(ctx, readinessErr).
						Msgf("failed to collect readiness; collector: %v", s.readiness)
					w.WriteHeader(http.StatusServiceUnavailable)

					_, writeErr = w.Write([]byte(notReadyMsg))

					return
				}

				switch {
				case readiness:
					s.Log().Trace(ctx).Msg("readiness inspection service reported that app IS ready")

					_, writeErr = w.Write([]byte("OK"))
				default:
					s.Log().Warn(ctx).Msg("readiness inspection service reported that app IS NOT ready yet")
					w.WriteHeader(http.StatusServiceUnavailable)

					_, writeErr = w.Write([]byte(notReadyMsg))
				}
			})
	}

	if s.facilities.Liveness() {
		mux.HandleFunc(
			"/"+s.livenessPathPrefix,
			func(w http.ResponseWriter, r *http.Request) {
				// TODO: now it is a mock. need to be implemented accordingly with real scenarios.
				_, writeErr := w.Write([]byte("OK"))
				if writeErr != nil {
					s.Log().
						Err(stopCtx, writeErr).
						Msgf("failed to write response for liveness probe request; request URI: %s", r.URL.Redacted())
				}
			})
	}

	if s.facilities.PProf() {
		mux.HandleFunc(fmt.Sprintf(`/%s/`, s.pprofPathPrefix), pprof.Index)
		mux.HandleFunc(fmt.Sprintf(`/%s/cmdline`, s.pprofPathPrefix), pprof.Cmdline)
		mux.HandleFunc(fmt.Sprintf(`/%s/profile`, s.pprofPathPrefix), pprof.Profile)
		mux.HandleFunc(fmt.Sprintf(`/%s/symbol`, s.pprofPathPrefix), pprof.Symbol)
		mux.HandleFunc(fmt.Sprintf(`/%s/trace`, s.pprofPathPrefix), pprof.Trace)
		mux.Handle(fmt.Sprintf(`/%s/goroutine`, s.pprofPathPrefix), pprof.Handler("goroutine"))
		mux.Handle(fmt.Sprintf(`/%s/heap`, s.pprofPathPrefix), pprof.Handler("heap"))
		mux.Handle(fmt.Sprintf(`/%s/allocs`, s.pprofPathPrefix), pprof.Handler("allocs"))
		mux.Handle(fmt.Sprintf(`/%s/block`, s.pprofPathPrefix), pprof.Handler("block"))
		mux.Handle(fmt.Sprintf(`/%s/mutex`, s.pprofPathPrefix), pprof.Handler("mutex"))
	}

	return
}

func makeMetricsHandler(stop context.Context, log logext.Logger) (res http.Handler, err error) {
	reg := prommetric.DefaultRegistry()

	biColl := collectors.NewBuildInfoCollector()
	if err = reg.Register(biColl); err != nil {
		return
	}

	goColl := collectors.NewGoCollector(
		collectors.WithGoCollectorMemStatsMetricsDisabled(),
		collectors.WithGoCollectorRuntimeMetrics(collectors.MetricsAll))
	if err = reg.Register(goColl); err != nil {
		return
	}

	res = promhttp.InstrumentMetricHandler(
		reg,
		promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				ErrorLog: func() promhttp.Logger {
					if log == nil || log.IsNop() {
						return nil
					}

					return &promHandlerErrorLogAdapter{log: log, logEvtCtx: stop}
				}(),
			}))

	return
}

func New(opts ...OptionFn) (res endpoint.Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of the HTTP diagnostics endpoint service; %w", err)
		}
	}()

	impl := &impl{}

	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	switch {
	case impl.livenessPathPrefix == "":
		err = errors.New("no liveness HTTP path prefix specified")

		return
	case impl.readinessPathPrefix == "":
		err = errors.New("no rediness HTTP path prefix specified")

		return
	case impl.pprofPathPrefix == "":
		err = errors.New("no pprof HTTP path prefix specified")

		return
	case impl.metricsPathPrefix == "":
		err = errors.New("no metrics HTTP path prefix specified")

		return
	}

	pathPrefixMap := map[string]struct{}{
		strings.ToLower(impl.metricsPathPrefix):   {},
		strings.ToLower(impl.livenessPathPrefix):  {},
		strings.ToLower(impl.readinessPathPrefix): {},
		strings.ToLower(impl.pprofPathPrefix):     {},
	}

	// nolint:revive,mnd
	if len(pathPrefixMap) != 4 {
		err = fmt.Errorf(
			"one of HTTP path prefixes not unique; paths: %s",
			strings.Join(
				[]string{
					impl.metricsPathPrefix,
					impl.livenessPathPrefix,
					impl.readinessPathPrefix,
					impl.pprofPathPrefix,
				},
				", "))

		return
	}

	switch {
	case impl.facilities.Readiness():
		if impl.readiness == nil {
			err = errors.New("no readiness inspection service specified")

			return
		}
	}

	res = impl

	return
}

func MustNew(opts ...OptionFn) endpoint.Service {
	res, err := New(opts...)
	if err != nil {
		panic(err)
	}

	return res
}
