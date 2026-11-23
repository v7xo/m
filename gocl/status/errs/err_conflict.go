package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrConflict = clerr.ErrStatusf(opstatus.Conflict, "entity state version conflicts with its current state")

func ConflictWrapWith(err error) error {
	return Wrap(ErrConflict, err)
}

func ConflictWrapWithf(err error, format string, a ...any) error {
	return Wrap(ErrConflict, fmt.Errorf(format, a...))
}
