package contextenrich

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Enabled(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.enabled = val })
}

func (o OptionFn) Conf(val Conf) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.enabled = val.Enabled })
}

func Option() OptionFn { return fopt.Nop[impl] }
