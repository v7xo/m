package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	underlying                http.RoundTripper
	clientName                string
	measureAsRetryableRequest bool
}

func (s *impl) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	startTs := time.Now().UTC()

	metricData := metricstub.HTTPRequestSecondsDataFromContext(req.Context())
	switch metricData {
	case nil:
		var reqAttrs httpext.RequestAttrs

		req, reqAttrs, err = httpext.RequestWithAttrs(req, false)
		if err != nil {
			return
		}

		metricData = metricstub.NewHTTPRequestSecondsData(
			s.clientName,
			reqAttrs.Method(),
			reqAttrs.URLAuthority(),
			reqAttrs.URLPath(),
			s.measureAsRetryableRequest,
		)
	default:
		metricData = metricstub.NewHTTPRequestSecondsData(
			s.clientName,
			metricData.Method(),
			metricData.URLAuthority(),
			metricData.URLPath(),
			s.measureAsRetryableRequest,
		)
	}

	req = req.WithContext(metricData.ToContext(req.Context()))
	resp, err = s.underlying.RoundTrip(req)
	durationSecs := time.Since(startTs).Seconds()

	switch err {
	case nil:
		if resp == nil {
			metricData.Status = placeholder.Nil
			metricData.StatusCode = placeholder.Nil
		} else {
			metricData.Status = http.StatusText(resp.StatusCode)
			metricData.StatusCode = strconv.Itoa(resp.StatusCode)
		}
	default:
		if resp == nil {
			metricData.StatusCode = placeholder.Failed
		} else {
			metricData.StatusCode = strconv.Itoa(resp.StatusCode)
		}

		metricData.Status = string(errs.TranslateStatus(err))
		metricData.Failed = true
	}

	metricData.Observe(durationSecs)

	return
}

func New(opts ...OptionFn) (res http.RoundTripper, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of retryable HTTP client metrics instrumenting; err: %w", err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.underlying == nil {
		err = errors.New("underlying round tripper not specified")

		return
	}

	if impl.clientName == "" {
		impl.clientName = placeholder.Unspecified
	}

	res = impl

	return
}
