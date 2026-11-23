package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrNotImplemented = clerr.ErrStatusf(opstatus.NotImplemented, "method or operation not implemented yet")

func NotImplementedWrapWith(err error) error {
	return Wrap(ErrNotImplemented, err)
}
