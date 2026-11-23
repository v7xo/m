package producer

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Name(val string) OptionFn { return fopt.Chain(o, func(t *impl) { t.name = val }) }

func (o OptionFn) Conf(val ClientConf) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.conf = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
