package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrAlreadyExists = clerr.ErrStatusLevelf(opst.AlreadyExists, level.Warn, "entity already exists")

func AlreadyExistsWrapWith(err error) error {
	return Wrap(ErrAlreadyExists, err)
}

func AlreadyExistsWrapWithf(err error, format string, a ...any) error {
	return Wrap(ErrAlreadyExists, fmt.Errorf(format, a...))
}
