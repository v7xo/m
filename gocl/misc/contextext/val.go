package contextext

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func Val[T any](ctx context.Context, key Key) (res T, ok bool, err error) {
	if ctx == nil {
		return
	}

	switch v := ctx.Value(key).(type) {
	case T:
		res, ok = v, true
	case nil:
		break
	default:
		err = fmt.Errorf(
			"such type of context's value is not expected; value key: %s; actual value type: %T; expected value type: %T; err: %w",
			key,
			v,
			res,
			errs.ErrUnexpectedTypeOfValueInContext)
	}

	return
}
