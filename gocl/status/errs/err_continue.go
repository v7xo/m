package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrConsumeLoopContinue = clerr.ErrStatusLevelf(opstatus.Unspecified, level.Warn, "consume loop continue ...")
