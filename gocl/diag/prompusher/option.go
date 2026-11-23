package prompusher

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(v logext.Logger) OptionFn { return fopt.WithLogScope(o, v, LogScope) }
func (o OptionFn) Conf(v Conf) OptionFn         { return fopt.Chain(o, func(t *impl) { t.conf = v }) }
func Option() OptionFn                          { return fopt.Nop[impl] }
