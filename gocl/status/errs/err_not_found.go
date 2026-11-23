package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrNotFound = clerr.ErrStatusLevelf(opstatus.NotFound, level.Warn, "not found")

func NotFoundWrapWith(err error) error {
	return Wrap(ErrNotFound, err)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
