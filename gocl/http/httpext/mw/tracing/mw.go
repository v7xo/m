package tracing

import (
	"net/http"

	"go.opentelemetry.io/otel/propagation"
)

// ParseTraceParentHeaders should be used if "traceparent" and "tracestate" fields provided in HEADERS.
func ParseTraceParentHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
}

// ParseBaggageHeaders should be used if baggage fields provided in HEADERS.
func ParseBaggageHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := propagation.Baggage{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
}
