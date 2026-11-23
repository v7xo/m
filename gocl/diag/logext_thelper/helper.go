package logext_thelper

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logfactory"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

//nolint:gochecknoglobals
var loggerSingleton2 = &atomic.Pointer[logext.Logger]{}

func Log(t *testing.T) logext.Logger {
	t.Helper()

	res := loggerSingleton2.Load()

	if res != nil {
		return *res
	}

	newLogger := NewLog(t)

	if !t.Failed() {
		newLoggerWin := loggerSingleton2.CompareAndSwap(nil, &newLogger)
		if newLoggerWin {
			newLogger.Info(t.Context()).
				Msgf("logger initialized and set; logger: %s", stringsext.TypeString(reflect.ValueOf(newLogger).Type()))

			return newLogger
		}

		return *loggerSingleton2.Load()
	}

	return logext.NopLogger
}

func NewLog(t *testing.T) logext.Logger {
	t.Helper()

	l, err := logfactory.Default()(nil)
	require.NoError(t, err, "unexpected error while creating a new logger")
	require.NotNil(t, l, "nil returned as a result of new logger factory function")

	return l
}

func FromContext(t *testing.T, ctx context.Context) logext.Logger {
	t.Helper()

	log, err := contextext.Log(ctx)
	require.NoError(t, err)
	require.NotNil(t, log)

	return log
}
