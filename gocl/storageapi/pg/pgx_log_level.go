package pg

import (
	pgxlog "github.com/jackc/pgx/v5/tracelog"

	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

func mapToPgxLogLevel(l level.Num) pgxlog.LogLevel {
	if l.IsInfo() {
		return pgxlog.LogLevelInfo
	}

	if l.IsWarn() {
		return pgxlog.LogLevelWarn
	}

	if l.IsError() {
		return pgxlog.LogLevelError
	}

	if l.IsTrace() {
		return pgxlog.LogLevelTrace
	}

	if l.IsDebug() {
		return pgxlog.LogLevelDebug
	}

	return DefaultFallbackPgxLogLevel
}
