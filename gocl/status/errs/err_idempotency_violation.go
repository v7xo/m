package errs

import (
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var (
	ErrIdempotencyViolation           = clerr.ErrStatusf(opst.IdempotencyViolation, "idempotency violation")
	ErrParamsLeadIdempotencyViolation = fmt.Errorf("params lead to violation of idempotency; err: %w", ErrIdempotencyViolation)
)

func IsIdempotencyViolation(err error) bool {
	return errors.Is(err, ErrIdempotencyViolation) || errors.Is(err, ErrParamsLeadIdempotencyViolation)
}
