package pg

import (
	"github.com/jackc/pgx/v5/tracelog"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

const (
	DefaultMinReadBufferSize   = 16384
	DefaultFallbackPgxLogLevel = tracelog.LogLevelInfo
	DefaultIsolationLevel      = storage.IsolationLevelRepeatableRead

	PgxConnLogOrigin = "pgx_conn"

	LogScope = "cl_storageapi_pg"
)
