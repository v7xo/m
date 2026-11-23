package pg

import (
	"context"
	"time"

	pgxlog "github.com/jackc/pgx/v5/tracelog"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
)

type (
	pgxLogAdapter struct {
		log logext.Logger
	}

	pgxLogDataExtractor func(e logext.Event, val any) logext.Event
)

//nolint:gochecknoglobals
var pgxLogDataExtractors = map[string]pgxLogDataExtractor{
	"host": func(e logext.Event, val any) logext.Event {
		if valAsStr, ok := val.(string); ok {
			return e.Str(attrname.Addr, valAsStr)
		}

		return e
	},
	"sql": func(e logext.Event, val any) logext.Event {
		if valAsStr, ok := val.(string); ok {
			return e.Str(attrname.StorageQueryText, valAsStr)
		}

		return e
	},
	"time": func(e logext.Event, val any) logext.Event {
		if valAsDur, ok := val.(time.Duration); ok {
			return e.Int64(attrname.StorageQueryTimeDurationMs, valAsDur.Milliseconds())
		}

		return e
	},
	"pid": func(e logext.Event, val any) logext.Event {
		if valAsUint32, ok := val.(uint32); ok {
			return e.Uint32(attrname.StorageClientBackendPid, valAsUint32)
		}

		return e
	},
	"err": func(e logext.Event, val any) logext.Event {
		if valAsErr, ok := val.(error); ok {
			return e.Err(valAsErr)
		}

		return e
	},
	"name": func(e logext.Event, val any) logext.Event {
		if valAsStr, ok := val.(string); ok {
			return e.Str(attrname.StoragePreparedStmtName, valAsStr)
		}

		return e
	},
	"rowCount": func(e logext.Event, val any) logext.Event {
		switch valStrict := val.(type) {
		case int64:
			return e.Int64(attrname.StorageRowCount, valStrict)
		case int32:
			return e.Int32(attrname.StorageRowCount, valStrict)
		case int:
			return e.Int(attrname.StorageRowCount, valStrict)
		case uint64:
			return e.Uint64(attrname.StorageRowCount, valStrict)
		case uint32:
			return e.Uint32(attrname.StorageRowCount, valStrict)
		case uint:
			return e.Uint(attrname.StorageRowCount, valStrict)
		default:
			return e
		}
	},
}

func (a *pgxLogAdapter) Log(ctx context.Context, l pgxlog.LogLevel, msg string, data map[string]any) {
	const (
		queryMsg   = "Query"
		prepareMsg = "Prepare"
		connectMsg = "Connect"
	)

	var evt logext.Event

	switch l {
	case pgxlog.LogLevelError:
		evt = a.log.Err(ctx, nil)
	case pgxlog.LogLevelWarn:
		evt = a.log.Warn(ctx)
	case pgxlog.LogLevelInfo:
		switch msg {
		case queryMsg, prepareMsg, connectMsg:
			// intentionally override log level.
			evt = a.log.Trace(ctx)
		default:
			evt = a.log.Info(ctx)
		}
	case pgxlog.LogLevelDebug, pgxlog.LogLevelNone, pgxlog.LogLevelTrace:
		evt = a.log.Trace(ctx)
	default:
		switch msg {
		case queryMsg, prepareMsg, connectMsg:
			// intentionally override log level.
			evt = a.log.Trace(ctx)
		default:
			evt = a.log.Info(ctx)
		}
	}

	evt = evt.Str(attrname.LogOrigin, PgxConnLogOrigin)

	evt = evt.WithSendMsg(msg)

	for k, v := range data {
		if extrFunc, ok := pgxLogDataExtractors[k]; ok {
			evt = extrFunc(evt, v)
		}
	}

	if queryExecProps, ok, _ := query.ExecPropsFromContext(ctx); ok {
		evt = evt.
			Str(attrname.StorageQueryName, queryExecProps.Name()).
			Int64(attrname.StorageQueryTimeoutMs, queryExecProps.Timeout().Milliseconds())
	}

	evt.Send()
}

func newTraceLogAdapter(log logext.Logger, overrideLogLevel level.Name) *pgxlog.TraceLog {
	if log == nil || log.IsNop() {
		return nil
	}

	logTune := log.Tune()
	level := log.LogLevel()

	if !overrideLogLevel.IsUnspecified() {
		overrideNum := overrideLogLevel.Num()

		if overrideNum < level {
			level = overrideNum
			logTune = logTune.Level(level)
		}
	}

	logTune = logTune.AddAttrs(attrname.Dict{attrname.LogScope: LogScope})

	return &pgxlog.TraceLog{
		LogLevel: mapToPgxLogLevel(level),
		Logger:   &pgxLogAdapter{log: logTune.Tuned()},
	}
}
