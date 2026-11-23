package subvol

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Findmnt(val findmnt.Findmnt) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.findmnt = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
