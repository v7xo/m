package filesystem

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Mount(val mount.Mount) OptionFn {
	return fopt.Chain(o, func(i *impl) { i.mount = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
