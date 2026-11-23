package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrStartRequested = clerr.Errf("unable to execute start operation; start has been already requested")
