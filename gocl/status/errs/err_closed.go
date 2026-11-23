package errs

import "rn.m.vo7.io/vo7io/m/gocl/status/clerr"

var ErrClosed = clerr.Errf("unable to execute operation; already closed")
