package contextext

import (
	"context"
)

func WithNoRateLimit(ctx context.Context, noRateLimit bool) context.Context {
	return context.WithValue(ctx, KeyNoRateLimit, noRateLimit)
}

func NoRateLimit(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	ctxVal := ctx.Value(KeyNoRateLimit)
	if ctxVal == nil {
		return false
	}

	if noRateLimit, ok := ctxVal.(bool); ok {
		return noRateLimit
	}

	return false
}
