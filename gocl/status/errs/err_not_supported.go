package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrNotSupported = clerr.Errf("method, operation or value not supported")

func NotSupportedWrap(err error) error { return Wrap(ErrNotSupported, err) }
