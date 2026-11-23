package errs

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrTypeMismatch = clerr.ErrStatusf(opstatus.InternalFail, "type mismatch")

func TypeMismatchFor[Expected any, Actual any](msgAndArgs ...any) error {
	var msgFull string

	if len(msgAndArgs) != 0 {
		msg := ""

		switch typed := msgAndArgs[0].(type) {
		case string:
			msg = typed
		default:
			msg = fmt.Sprintf("%v", msgAndArgs[0])
		}

		if len(msgAndArgs) != 1 {
			msgFull = fmt.Sprintf(msg, msgAndArgs[1:])
		} else {
			msgFull = msg
		}
	}

	if msgFull == "" {
		return fmt.Errorf("%w; expected type '%s', actual '%s'", ErrTypeMismatch, typestring.TypeStringFor[Expected](), typestring.TypeStringFor[Actual]())
	}

	return fmt.Errorf("%w; expected type '%s', actual '%s'; %s", ErrTypeMismatch, typestring.TypeStringFor[Expected](), typestring.TypeStringFor[Actual](), msgFull)
}
