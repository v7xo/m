package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrAlreadyLocked = clerr.ErrStatusf(opstatus.AlreadyLocked, "lock already held")

func AlreadyLockedWrapWith(err error) error {
	return Wrap(ErrAlreadyLocked, err)
}

func AlreadyLockedWrapWithf(err error, format string, a ...any) error {
	return Wrap(ErrAlreadyLocked, fmt.Errorf(format, a...))
}
