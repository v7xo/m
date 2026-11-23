package value

import "rn.m.vo7.io/vo7io/m/gocl/status/errs"

func Convert[From any, To any](in From, msgAndArgs ...any) (res To, err error) {
	typed, ok := any(in).(To)
	if !ok {
		return res, errs.TypeMismatchFor[To, From](msgAndArgs...)
	}

	return typed, nil
}
