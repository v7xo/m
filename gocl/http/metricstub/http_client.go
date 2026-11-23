package metricstub

import (
	"context"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// HTTPRequestSecondsData represents data for [HTTPRetryableRequestSeconds] and [HTTPRequestSeconds] metrics.
type HTTPRequestSecondsData struct {
	APIStatusCode     string
	clientName        string
	method            string
	Status            string
	StatusCode        string
	urlAuthority      string
	urlPath           string
	Retried           bool
	Failed            bool
	retryable         bool
	constructorCalled bool
}

func NewHTTPRequestSecondsData(
	clientName string,
	method string,
	urlAuthority string,
	urlPath string,
	retryable bool,
) *HTTPRequestSecondsData {
	clientName = placeholder.UnspecifiedIfEmpty(clientName)
	method = placeholder.UnspecifiedIfEmpty(method)
	urlAuthority = placeholder.UnspecifiedIfEmpty(urlAuthority)
	urlPath = placeholder.UnspecifiedIfEmpty(urlPath)

	return &HTTPRequestSecondsData{
		clientName:   clientName,
		method:       method,
		retryable:    retryable,
		urlAuthority: urlAuthority,
		urlPath:      urlPath,

		APIStatusCode: placeholder.NotAvailable,

		constructorCalled: true,
	}
}

func (v *HTTPRequestSecondsData) Retryable() bool {
	if v == nil {
		return false
	}

	return v.retryable
}

func (v *HTTPRequestSecondsData) URLAuthority() string {
	if v == nil {
		return ""
	}

	return v.urlAuthority
}

func (v *HTTPRequestSecondsData) URLPath() string {
	if v == nil {
		return ""
	}

	return v.urlPath
}

func (v *HTTPRequestSecondsData) Method() string {
	if v == nil {
		return ""
	}

	return v.method
}

func (v *HTTPRequestSecondsData) ClientName() string {
	if v == nil {
		return ""
	}

	return v.clientName
}

func (v *HTTPRequestSecondsData) ToContext(ctx context.Context) context.Context {
	if v == nil {
		return ctx
	} else if !v.constructorCalled {
		panic(errs.ErrUninitializedValue)
	}

	return context.WithValue(ctx, contextext.KeyHTTPRequestSecondsData, v)
}

func HTTPRequestSecondsDataFromContext(ctx context.Context) *HTTPRequestSecondsData {
	res, ok, err := contextext.Val[*HTTPRequestSecondsData](ctx, contextext.KeyHTTPRequestSecondsData)
	if err != nil || !ok {
		return nil
	}

	return res
}

func (v *HTTPRequestSecondsData) Observe(durationSecs float64) {
	if v == nil {
		return
	} else if !v.constructorCalled {
		panic(errs.ErrUninitializedValue)
	}

	if v.retryable {
		HTTPRetryableRequestSeconds.
			WithLabelVals(
				v.clientName,
				v.method,
				v.urlAuthority,
				v.urlPath,
				v.Status,
				v.StatusCode,
				v.APIStatusCode,
				strconv.FormatBool(v.Retried),
				strconv.FormatBool(v.Failed),
			).
			Observe(durationSecs)

		return
	}

	HTTPRequestSeconds.
		WithLabelVals(
			v.clientName,
			v.method,
			v.urlAuthority,
			v.urlPath,
			v.Status,
			v.StatusCode,
			v.APIStatusCode,
			strconv.FormatBool(v.Failed),
		).
		Observe(durationSecs)
}

var (
	HTTPRetryableRequestSeconds = prommetric.MustNewHistogramVec(
		false,
		prommetric.CLMetricsSubsystemName("http_client"),
		"retryable_request_seconds2",
		"indicates the count/duration of HTTP request issued by specified client; duration accumulates also time of retries, if any",
		[]string{
			string(attrname.ClientName),
			string(attrname.HTTPReqMethod),
			string(attrname.HTTPReqURLAuthority),
			string(attrname.HTTPReqURLPath),
			string(attrname.Status),
			string(attrname.StatusCode),
			string(attrname.APIStatusCode),
			string(attrname.Retried),
			string(attrname.HasFailed),
		},
	)

	HTTPRequestSeconds = prommetric.MustNewHistogramVec(
		false,
		prommetric.CLMetricsSubsystemName("http_client"),
		"request_seconds2",
		"indicates the count/duration of HTTP request issued by specified client",
		[]string{
			string(attrname.ClientName),
			string(attrname.HTTPReqMethod),
			string(attrname.HTTPReqURLAuthority),
			string(attrname.HTTPReqURLPath),
			string(attrname.Status),
			string(attrname.StatusCode),
			string(attrname.APIStatusCode),
			string(attrname.HasFailed),
		},
	)
)
