package contextext

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext/logtypes"
)

func MustLog(ctx context.Context) logtypes.Logger {
	res, err := Log(ctx)
	if err != nil {
		panic(err)
	}

	return res
}

func Log(ctx context.Context) (logtypes.Logger, error) {
	l, ok, err := Val[logtypes.Logger](ctx, KeyLog)
	if err != nil {
		return logtypes.NopLogger, err
	}

	if !ok || l == nil {
		return logtypes.NopLogger, nil
	}

	return l, nil
}

func LogExists(ctx context.Context) (logtypes.Logger, bool, error) {
	l, ok, err := Val[logtypes.Logger](ctx, KeyLog)
	if err != nil || !ok {
		return logtypes.NopLogger, false, err
	}

	if l == nil {
		return logtypes.NopLogger, true, nil
	}

	return l, true, nil
}

func WithLog(ctx context.Context, log logtypes.Logger) context.Context {
	return context.WithValue(ctx, KeyLog, log)
}
