package opstatus

import (
	"net/http"
)

var (
	fallbackHTTPStatus = http.StatusInternalServerError

	toHTTPStatusMap = map[Name]int{
		AlreadyExists:              http.StatusConflict,
		AlreadyLocked:              http.StatusPreconditionFailed,
		AmbiguousMatch:             http.StatusConflict,
		BadArgs:                    http.StatusBadRequest,
		CallerCancel:               http.StatusRequestTimeout,
		CallerRateLimit:            http.StatusTooManyRequests,
		CallerTimeout:              http.StatusRequestTimeout,
		Cancel:                     http.StatusRequestTimeout,
		Conflict:                   http.StatusConflict,
		DataContractInvalidVersion: http.StatusUnprocessableEntity,
		DataContractViolation:      http.StatusUnprocessableEntity,
		DBConstraintViolation:      http.StatusUnprocessableEntity,
		DBFail:                     http.StatusInternalServerError,
		Fail:                       http.StatusInternalServerError,
		IdempotencyViolation:       http.StatusPreconditionFailed,
		InternalFail:               http.StatusInternalServerError,
		NotFound:                   http.StatusNotFound,
		NotImplemented:             http.StatusNotImplemented,
		OK:                         http.StatusOK,
		Timeout:                    http.StatusRequestTimeout,
		TooManyRequests:            http.StatusTooManyRequests,
		Unauthned:                  http.StatusUnauthorized,
		Unauthzed:                  http.StatusForbidden,
		UnknownStatus:              fallbackHTTPStatus,
		Unspecified:                fallbackHTTPStatus,
	}
)

func (v Name) HTTPStatus() int {
	if inMap, ok := toHTTPStatusMap[v]; ok {
		return inMap
	}

	return fallbackHTTPStatus
}
