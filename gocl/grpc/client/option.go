package client

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/client/echo"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn[TClient echo.Echo] func(*factory[TClient]) error

func (o OptionFn[TClient]) ClientConstructor(val ClientConstructorFn[TClient]) OptionFn[TClient] {
	return fopt.Chain(o, func(t *factory[TClient]) { t.clientConstructor = val })
}

func (o OptionFn[TClient]) Conf(val Conf) OptionFn[TClient] {
	return fopt.Chain(o, func(t *factory[TClient]) { t.conf = val })
}

func (o OptionFn[TClient]) Log(v logext.Logger) OptionFn[TClient] {
	return fopt.WithLogScope(o, v, LogScope)
}

func Option[TClient echo.Echo]() OptionFn[TClient] { return fopt.Nop[factory[TClient]] }
