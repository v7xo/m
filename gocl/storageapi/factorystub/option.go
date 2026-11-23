package factorystub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	factorypkg "rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
)

type OptionFn[TAPI storageapi.SQLAPI] func(*impl[TAPI]) error

func (o OptionFn[TAPI]) Log(val logext.Logger) OptionFn[TAPI] {
	return fopt.WithLogScope(o, val, LogScope)
}

func (o OptionFn[TAPI]) ClientConf(val ClientConf) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.clientConf = val })
}

func (o OptionFn[TAPI]) InstanceContextKey(val contextext.Key) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.instanceContextKey = val })
}

func (o OptionFn[TAPI]) APIConstructor(val factorypkg.APIConstructorFn[TAPI]) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.apiConstructor = val })
}

func Option[TAPI storageapi.SQLAPI]() OptionFn[TAPI] { return fopt.Nop[impl[TAPI]] }
