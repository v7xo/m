package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrAmbiguousMatch = clerr.ErrStatusf(opstatus.AmbiguousMatch, "a lot of entities match to condition")

func AmbiguousMatchWrapWith(err error) error {
	return Wrap(ErrAmbiguousMatch, err)
}

func AmbiguousMatchWrapWithf(err error, format string, a ...any) error {
	return Wrap(ErrAmbiguousMatch, fmt.Errorf(format, a...))
}
