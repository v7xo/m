package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrNotAllowedMediaType = clerr.ErrStatusf(opstatus.BadArgs, "specified media type is either not supported or not allowed")

func NotAllowedMediaTypeWrap(with error) error {
	return Wrap(ErrNotAllowedMediaType, with)
}

func NotAllowedMediaTypeWrapf(format string, a ...any) error {
	return Wrap(ErrNotAllowedMediaType, fmt.Errorf(format, a...))
}
