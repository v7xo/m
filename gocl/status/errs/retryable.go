package errs

import (
	"errors"
)

func RequestRetryable(err error) bool {
	var timeoutError interface{ Timeout() bool }
	if errors.As(err, &timeoutError) && timeoutError.Timeout() {
		return false
	}

	var tempError interface{ Temporary() bool }
	if errors.As(err, &tempError) {
		return tempError.Temporary()
	}

	return false
}
