package errs

import (
	"fmt"
	"runtime/debug"
)

// PanicErrUnconditionally translates panic object p to error unconditionally.
// Unconditionally means that func will return error even p is nil.
func PanicErrUnconditionally(p any) error {
	panicErr, isPanicErr := p.(error)
	if isPanicErr {
		panicErr = fmt.Errorf("panic: %w; stack: %s", panicErr, string(debug.Stack()))
	} else {
		panicErr = fmt.Errorf("panic: %v; stack: %s", p, string(debug.Stack()))
	}

	return panicErr
}

// PanicErr translates panic object p to error.
// If p is nil then nil returned.
func PanicErr(p any) error {
	if p == nil {
		return nil
	}

	panicErr, isPanicErr := p.(error)
	if isPanicErr {
		panicErr = fmt.Errorf("panic: %w; stack: %s", panicErr, string(debug.Stack()))
	} else {
		panicErr = fmt.Errorf("panic: %v; stack: %s", p, string(debug.Stack()))
	}

	return panicErr
}
