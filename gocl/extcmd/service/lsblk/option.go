package lsblk

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(log logext.Logger) OptionFn { return fopt.WithLogScope(o, log, LogScope) }

func Option() OptionFn { return fopt.Nop[impl] }
