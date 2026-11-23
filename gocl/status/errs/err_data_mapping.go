package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrDataMapping = clerr.ErrStatusf(opst.InternalFail, "data mapping error")
