package storage

import (
	"context"

	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

const (
	base10    = 10
	bitSize64 = 64
)

type OptionFn func(*impl) error

func (o OptionFn) StorageAPI(val storageapi.SQLAPI) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.storageAPI = val })
}

func (o OptionFn) Log(val logging.Logger) OptionFn {
	return fopt.WithLogScope(o, val, LogScope)
}

func (o OptionFn) FallbackContext(val context.Context) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.fallbackCtx = val })
}

func (o OptionFn) MetaInfoSchemaName(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.metaInfoSchemaName = val })
}

func (o OptionFn) MetaInfoVersionTableName(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.metaInfoVersionTableName = val })
}

func (o OptionFn) MetaInfoDeployTableName(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.metaInfoDeployTableName = val })
}

func (o OptionFn) LockID(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.lockID = val })
}

func (o OptionFn) DeployFingerprint(val types.DeployFingerprint) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.currentDeployFingerprint = int64(val) })
}

func (o OptionFn) StepRunTimeout(val timeout.Timeout) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.stepRunTimeout = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
