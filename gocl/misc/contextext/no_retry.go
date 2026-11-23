package contextext

import (
	"context"
)

func WithNoRetry(ctx context.Context, noRetry bool) context.Context {
	return context.WithValue(ctx, KeyNoRetry, noRetry)
}

func NoRetry(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	ctxVal := ctx.Value(KeyNoRetry)
	if ctxVal == nil {
		return false
	}

	if noRetry, ok := ctxVal.(bool); ok {
		return noRetry
	}

	return false
}
