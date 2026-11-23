package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrBadArgs = clerr.ErrStatusf(opstatus.BadArgs, "bad args")

func BadArgsWrapWith(err error) error {
	if errors.Is(err, ErrBadArgs) {
		return err
	}

	return Wrap(ErrBadArgs, err)
}
