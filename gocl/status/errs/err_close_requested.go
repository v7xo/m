package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrCloseRequested = clerr.Errf("unable to execute operation; close has been requested")

func IsCloseRequested(err error) bool {
	return errors.Is(err, ErrCloseRequested)
}
