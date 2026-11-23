package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrNotLocked = clerr.ErrStatusf(opstatus.Fail, "lock not held")

func NotLockedWrapWith(err error) error {
	return Wrap(ErrNotLocked, err)
}

func NotLockedWrapWithf(err error, format string, a ...any) error {
	return Wrap(ErrNotLocked, fmt.Errorf(format, a...))
}
