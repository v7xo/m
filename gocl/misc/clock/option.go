package clock

import (
	"context"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLog(o, val) }

func (o OptionFn) NowFn(val func(context.Context) time.Time) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.nowFn = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
