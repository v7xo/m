package logutil

import (
	"runtime"
	"strconv"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

const (
	slashSeparator = '/'
)

func CallerFnFile(depth int) (fnName string, fileLine string) {
	fnName, fileLine = placeholder.Failed, placeholder.Failed

	progCnt, file, line, ok := runtime.Caller(depth)
	if !ok {
		return
	}

	idx := strings.LastIndexByte(file, slashSeparator)
	fileLine = file[idx+1:] + ":" + strconv.Itoa(line)

	if fn := runtime.FuncForPC(progCnt); fn != nil {
		fnName = fn.Name()

		idx = strings.LastIndexByte(fnName, slashSeparator)
		if idx > -1 {
			fnName = strings.Replace(fnName[idx+1:], ".", "/", 1)
		}
	}

	return
}
