package metrics

import (
	"net/http"
	"strconv"
	"time"

	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw"
	"rn.m.vo7.io/vo7io/m/gocl/http/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type responseWriterWrapper struct {
	underlying http.ResponseWriter

	writeStarted bool
	statusCode   int
}

var _ http.ResponseWriter = &responseWriterWrapper{}

func (s *responseWriterWrapper) Header() http.Header {
	return s.underlying.Header()
}

func (s *responseWriterWrapper) Write(b []byte) (n int, err error) {
	n, err = s.underlying.Write(b)

	if !s.writeStarted {
		s.writeStarted = true
		s.statusCode = http.StatusOK
	}

	return
}

func (s *responseWriterWrapper) WriteHeader(statusCode int) {
	s.underlying.WriteHeader(statusCode)

	if !s.writeStarted {
		s.writeStarted = true
		s.statusCode = statusCode
	}
}

func New(serverName string) mw.Func {
	return func(next http.Handler) http.Handler {
		hnd := func(w http.ResponseWriter, r *http.Request) {
			startTs := time.Now().UTC()
			rww := &responseWriterWrapper{underlying: w}
			method, urlPath := "", ""

			enrichedReq, reqAttrs, reqAttrsErr := httpext.RequestWithAttrs(r, false)
			if reqAttrsErr != nil {
				method, _, urlPath = httpext.RequestMethodURLMetricLabels(r)
			} else {
				r = enrichedReq
				method, urlPath = reqAttrs.Method(), reqAttrs.URLPath()
			}

			next.ServeHTTP(rww, r)

			dur := time.Since(startTs).Seconds()
			statusCodeString := strconv.Itoa(rww.statusCode)

			metricstub.HTTPServerReqDurSeconds.
				WithLabelVals(
					placeholder.UnspecifiedIfEmpty(serverName),
					method,
					urlPath,
					http.StatusText(rww.statusCode),
					statusCodeString,
				).
				Observe(dur)
		}

		return http.HandlerFunc(hnd)
	}
}
