package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrTooManyRequests = clerr.ErrStatusf(opst.TooManyRequests, "too many requests")

func TooManyRequestsWrap(err error) error { return Wrap(ErrTooManyRequests, err) }
func IsTooManyRequests(err error) bool    { return errors.Is(err, ErrTooManyRequests) }
