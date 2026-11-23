package util

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

type AppReleaseFn func() string

//nolint:gochecknoglobals
var (
	appReleaseFunc   AppReleaseFn = info.UserAgentLong
	appReleaseFuncMu sync.RWMutex
)

func CreateStorageLockIDFromStorageConnConfig(cfg storage.ConnConf) string {
	return CreateStorageLockIDFromStorageName(cfg.Database)
}

func CreateStorageLockIDFromStorageName(name string) string {
	return name + "#colib-db-migration-lock"
}

func SetAppReleaseFn(f AppReleaseFn) {
	if f == nil {
		panic(errs.BadArgsWrapWith(errors.New("deploy app release function can't be nil")))
	}

	appReleaseFuncMu.Lock()
	appReleaseFunc = f
	appReleaseFuncMu.Unlock()
}

func AppRelease() string {
	appReleaseFuncMu.RLock()

	f := appReleaseFunc

	appReleaseFuncMu.RUnlock()

	return f()
}

func FormatVersionNums(v []uint) string {
	if len(v) == 0 {
		return placeholder.Empty
	}

	strs := make([]string, len(v))

	for i := range v {
		strs[i] = strconv.FormatUint(uint64(v[i]), 10) //nolint:revive,mnd
	}

	return strings.Join(strs, ", ")
}
