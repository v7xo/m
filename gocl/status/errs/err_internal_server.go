package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrInternalServerError = clerr.ErrStatusLevelf(opstatus.InternalFail, level.Error, "internal server error")
