package errs

import (
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
)

func Joinf(a error, msg string, args ...any) error {
	if ptrof.IsNil(a) {
		return fmt.Errorf(msg, args...)
	}

	b := fmt.Errorf(msg, args...)

	return fmt.Errorf("%w; %w", a, b)
}

func Join(e ...error) error {
	var res error

	for i := range len(e) {
		if ptrof.IsNil(e[i]) {
			continue
		}

		if res == nil {
			res = e[i]

			continue
		}

		res = fmt.Errorf("%w; %w", res, e[i])
	}

	return res
}

// JoinIfNot joins specified target error if err is not target (see [errors.Is]).
func JoinIfNot(err error, target error) error {
	if ptrof.IsNil(target) || ptrof.IsNil(err) || errors.Is(err, target) {
		return err
	}

	return fmt.Errorf("%w; %w", target, err)
}
