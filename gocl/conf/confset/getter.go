package confset

import (
	"fmt"
	"reflect"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type ConfGetter struct{ set *set }

func (c ConfGetter) Require[T flenv.Conf]() (T, error) {
	var (
		zero T
		typ  = reflect.TypeFor[T]()
	)

	res, ok := c.set.confs[typ]
	if !ok {
		return zero, fmt.Errorf("%w; no configuration of type '%s' found in configuration set", errs.ErrNotFound, typestring.TypeStringFor[T]())
	}

	switch resTyped := res.(type) {
	case T:
		return resTyped, nil
	default:
		return zero,
			fmt.Errorf("%w; type '%s' of configuration found in configuration set is not compatible with required '%s'",
				errs.ErrTypeMismatch,
				typestring.TypeStringOfVal(res),
				typestring.TypeStringFor[T]())
	}
}

func (c ConfGetter) MustRequire[T flenv.Conf]() T {
	res, err := c.Require[T]()
	if err != nil {
		panic(err)
	}

	return res
}
