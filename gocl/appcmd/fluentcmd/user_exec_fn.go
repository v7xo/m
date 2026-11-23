package fluentcmd

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type UserExecFn[T flenv.Conf] func(context.Context, T, logext.Logger) error
