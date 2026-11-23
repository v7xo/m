package errs

import "errors"

var ErrMissingValue = errors.New("value is missing and wasn't explicitly assigned")
