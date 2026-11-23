package contextext

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

// [Key] represents unified type of value denoting context scoped value key.
type Key = attrname.ContextKey

const (
	KeyLogAttrs = attrname.ContextKeyLogAttrs

	KeyCorrelationID          = Key(attrname.CorrelationID)
	KeyCurrentSession         = Key("current_session")
	KeyCurrentUser            = Key("current_user")
	KeyGoTestName             = Key(attrname.GoTestName)
	KeyHTTPRequestLogAttrs    = Key("http_req_log_attrs")
	KeyHTTPRequestSecondsData = Key("http_req_secs_data")
	KeyLog                    = Key(attrname.Log) // [KeyLog] denotes a key to associate logger with context.
	KeyNoRateLimit            = Key(attrname.NoRateLimit)
	KeyNoRetry                = Key(attrname.NoRetry)
	KeyRequestID              = Key(attrname.RequestID)
	KeySessionID              = Key(attrname.SessionID)
	KeyStorageAPIInstance     = Key(placeholder.Default)
	KeyStorageQueryExecProps  = Key(attrname.StorageQueryExecProps)
	KeyStorageQueryExecStats  = Key(attrname.StorageQueryExecStats)
	KeyTraceIDAlt             = Key(attrname.TraceIDAlt)
	KeyTraceParent            = Key(attrname.TraceParent)
	KeyUserID                 = Key(attrname.UserID)
	KeyUserRole               = Key(attrname.UserRole)
)
