package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var (
	ErrLogged          = errors.New("<error sent to log>")
	ErrLogUnknownLevel = clerr.Errf("unknown log level")
)
