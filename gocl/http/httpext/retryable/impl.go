package retryable

import (
	"fmt"
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	clhttp "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/metrics"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/retrier"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/singler"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/tracing"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type impl struct {
	fopt.LogOpt
	retrierOpts     retrier.OptionFn
	singlerOpts     singler.OptionFn
	tracingOpts     tracing.OptionFn
	pipeline        http.RoundTripper
	observeResultFn ObserveResultFn
	clientName      string
	metricsEnabled  bool
	tracingEnabled  bool
	retryable       bool
}

// RoundTrip implements [http.RoundTripper] interface.
func (s *impl) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if s.observeResultFn != nil {
		defer func() {
			s.observeResultFn(req, resp, err)
		}()
	}

	req, _, err = clhttp.RequestWithAttrs(req, true)
	if err != nil {
		return
	}

	resp, err = s.pipeline.RoundTrip(req)

	return
}

func (s *impl) initLog() error {
	return s.SetLog(
		s.Log().
			Tune().
			AddAttrs(attrname.Dict{attrname.APIClientName: placeholder.UnspecifiedIfEmpty(s.clientName)}).
			Tuned(),
	)
}

func New(opts ...OptionFn) (_ *http.Client, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{retrierOpts: retrier.Option(), singlerOpts: singler.Option(), tracingOpts: tracing.Option()}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if err = impl.initLog(); err != nil {
		return nil, err
	}

	impl.pipeline, err = singler.New(impl.singlerOpts, singler.Option().Log(impl.Log()))
	if err != nil {
		return nil, err
	}

	if impl.tracingEnabled {
		impl.pipeline, err = tracing.New(
			impl.tracingOpts,
			tracing.Option().
				Underlying(impl.pipeline).
				Log(impl.Log()),
		)
		if err != nil {
			return nil, err
		}
	}

	if impl.retryable && impl.metricsEnabled {
		// measures single round trip (if non-retryable then all requests are measured with mw below by code).
		impl.pipeline, err = metrics.New(
			metrics.Option().
				ClientName(impl.clientName).
				RetryableRequest(false).
				Underlying(impl.pipeline),
		)
		if err != nil {
			return nil, err
		}
	}

	if impl.retryable {
		impl.pipeline, err = retrier.New(
			impl.retrierOpts,
			retrier.Option().
				Underlying(impl.pipeline).
				Log(impl.Log()),
		)
		if err != nil {
			return nil, err
		}
	}

	if impl.metricsEnabled {
		impl.pipeline, err = metrics.New(
			metrics.Option().
				ClientName(impl.clientName).
				RetryableRequest(impl.retryable).
				Underlying(impl.pipeline),
		)
		if err != nil {
			return nil, err
		}
	}

	return &http.Client{Transport: impl}, nil
}
