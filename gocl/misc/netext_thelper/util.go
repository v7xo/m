package netext_thelper

import (
	"errors"
	"runtime"
	"strings"
	"syscall"
)

func isErrAddrAlreadyInUse(err error) bool {
	if syscallErrno, ok := errors.AsType[syscall.Errno](err); ok {
		switch strings.ToUpper(runtime.GOOS) {
		case "WINDOWS":
			return syscallErrno == windowsAddrInUserErrno
		default:
			return syscallErrno == syscall.EADDRINUSE
		}
	}

	return false
}
