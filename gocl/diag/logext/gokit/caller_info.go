package gokit

import (
	lgu "rn.m.vo7.io/vo7io/m/gocl/diag/logutil"
)

type callerInfoFn func() (fnName, fileLine string)

func makeCallerInfo(depth int) callerInfoFn {
	return func() (fnName, fileLine string) {
		return lgu.CallerFnFile(depth)
	}
}
