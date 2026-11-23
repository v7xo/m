package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrUnexpectedTypeOfValueInContext = clerr.ErrStatusf(opstatus.InternalFail, "unexpected type of value within context")
