package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrInvalidOperation = clerr.ErrStatusf(opst.Fail, "invalid operation")
