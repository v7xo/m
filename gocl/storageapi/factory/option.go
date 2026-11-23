package factory

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type OptionFn[TAPI storageapi.SQLAPI] func(*impl[TAPI]) error

func (o OptionFn[TAPI]) Log(val logext.Logger) OptionFn[TAPI] {
	return fopt.WithLogScope(o, val, LogScope)
}

func (o OptionFn[TAPI]) InstanceContextKey(val contextext.Key) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.instanceContextKey = val })
}

func (o OptionFn[TAPI]) ConnConf(val storage.ConnConf) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.connConf = val })
}

func (o OptionFn[TAPI]) APIConf(val storageapi.Conf) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.apiConf = val })
}

func (o OptionFn[TAPI]) APIConstructor(val APIConstructorFn[TAPI]) OptionFn[TAPI] {
	return fopt.Chain(o, func(t *impl[TAPI]) { t.apiConstructor = val })
}

func FlenvRootOption[TAPI storageapi.SQLAPI]() OptionFn[TAPI] {
	var opt OptionFn[TAPI] = func(fact *impl[TAPI]) (err error) {
		connConf, err := storage.FlenvRootConnConf()
		if err != nil {
			return
		}

		apiConf, err := storageapi.FlenvRootConf()
		if err != nil {
			return
		}

		fact.connConf = connConf
		fact.apiConf = apiConf

		return nil
	}

	return opt
}

func Option[TAPI storageapi.SQLAPI]() OptionFn[TAPI] { return fopt.Nop[impl[TAPI]] }

type OptionSQLAPIFn = OptionFn[storageapi.SQLAPI]

func OptionSQLAPI() OptionSQLAPIFn { return Option[storageapi.SQLAPI]() }
