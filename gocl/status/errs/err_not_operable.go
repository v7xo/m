package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrNotOperable = clerr.ErrStatusf(opst.InternalFail, "unable to execute operation; not operable")

func IsNotOperable(err error) bool {
	return errors.Is(err, ErrNotOperable)
}
