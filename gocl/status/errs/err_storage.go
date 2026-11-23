package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var (
	ErrConstraintViolation = clerr.ErrStatusf(opstatus.DBConstraintViolation, "operation leads to storage constraint violation")

	ErrStorageTypeUnsupported = clerr.ErrStatusf(opstatus.DBFail, "storage type is not supported")
	ErrNoStorageAPIInContext  = clerr.ErrStatusf(opstatus.DBFail, "current context not associated with required storage API instance")

	ErrTxExists               = clerr.ErrStatusf(opstatus.DBFail, "tx has already been opened")
	ErrTxRequired             = clerr.ErrStatusf(opstatus.DBFail, "tx required to perform this operation")
	ErrTxInconsistentRollback = clerr.ErrStatusf(opstatus.DBFail, "inconsistent tx rollback operation")
	ErrTxInconsistentCommit   = clerr.ErrStatusf(opstatus.DBFail, "inconsistent tx commit operation")
	ErrTxRollingback          = clerr.ErrStatusf(opstatus.DBFail, "tx rolling back")

	// ErrSerializationFailure corresponds to name from postgres documentation (serialization failure).
	ErrTxSerializationFailure = clerr.ErrStatusf(opstatus.DBFail, "tx serialization failure")

	ErrNamedQueryParamMissing = clerr.ErrStatusf(opstatus.DBFail, "named query parameter is missing")
	ErrInvalidQueryParamName  = clerr.ErrStatusf(opstatus.DBFail, "name of query parameter is invalid")
)
