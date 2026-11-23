package contextext

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func SpanIDString(ctx context.Context) string {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if spanContext.IsValid() {
		return spanContext.SpanID().String()
	}

	return ""
}

func TraceIDString(ctx context.Context) string {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		if traceIDAlt, ok, _ := TraceIDAlt(ctx); ok {
			return traceIDAlt
		}

		return ""
	}

	return spanContext.TraceID().String()
}

func WithTraceIDAlt(ctx context.Context, traceIDAlt string) context.Context {
	return context.WithValue(ctx, KeyTraceIDAlt, traceIDAlt)
}

func TraceIDAlt(ctx context.Context) (string, bool, error) {
	return Val[string](ctx, KeyTraceIDAlt)
}
