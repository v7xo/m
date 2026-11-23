package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrUnidentifiedValue = clerr.ErrStatusf(opstatus.InternalFail, "unidentified value")
