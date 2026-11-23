package gomigrate

import (
	"github.com/golang-migrate/migrate/v4/source"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/drivers"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) SourceGoTemplatingEnabled(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.sourceGoTemplatingEnabled = val })
}

func (o OptionFn) StorageDriver(n types.StorageDriverName, d drivers.StorageDriver) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.storageDriver = d; t.storageDriverName = n })
}

func (o OptionFn) SourceDriver(n types.SourceDriverName, d source.Driver) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.sourceDriver = d; t.sourceDriverName = n })
}

func (o OptionFn) DeployJobID(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.deployJobID = val })
}

func Option() OptionFn { return func(*impl) error { return nil } }
