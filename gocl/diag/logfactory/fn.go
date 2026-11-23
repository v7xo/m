package logfactory

import "rn.m.vo7.io/vo7io/m/gocl/diag/logext"

type Fn func(logext.ConfGetter) (logext.Logger, error)
