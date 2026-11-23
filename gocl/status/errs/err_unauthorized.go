package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var (
	ErrUnauthorized = clerr.ErrStatusf(opst.Unauthzed, "unauthorized access")
	ErrForbidden    = ErrUnauthorized
)

func UnauthorizedWrap(err error) error { return Wrap(ErrUnauthorized, err) }
