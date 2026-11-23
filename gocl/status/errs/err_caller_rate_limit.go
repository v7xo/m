package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrCallerRateLimit = clerr.ErrStatusf(opstatus.CallerRateLimit, "request canceled on caller side due to rate limit occurs")

func CallerRateLimitWrapWith(err error) error {
	if errors.Is(err, ErrCallerRateLimit) {
		return err
	}

	return Wrap(ErrCallerRateLimit, err)
}
