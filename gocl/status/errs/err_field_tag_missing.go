package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrFieldTagMissing = clerr.Errf("no field annotated with tag")
