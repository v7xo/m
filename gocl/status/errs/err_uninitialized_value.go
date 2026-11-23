package errs

import "errors"

var (
	ErrUninitializedValue  = errors.New("value wasn't initialized")
	ErrNotInitializedValue = ErrUninitializedValue
)
