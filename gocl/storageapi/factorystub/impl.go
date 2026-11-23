package factorystub

import (
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	factorypkg "rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type impl[TAPI storageapi.SQLAPI] struct {
	fopt.LogOpt
	factorypkg.Factory[TAPI]
	apiConstructor     factorypkg.APIConstructorFn[TAPI]
	instanceContextKey contextext.Key
	clientConf         storageapi.ClientConf
}

func New[TAPI storageapi.SQLAPI](opts ...OptionFn[TAPI]) (_ Factory[TAPI], err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl[TAPI]](), err)
		}
	}()

	f := &impl[TAPI]{instanceContextKey: InstanceContextKey}
	if err = fopt.Apply(f, opts...); err != nil {
		return nil, err
	}

	if err = f.clientConf.Validate(validation.DefaultValidator()); err != nil {
		return nil, fmt.Errorf("invalid client configuration supplied; %w", err)
	}

	if f.apiConstructor == nil {
		return nil, errors.New("no API constructor function specified")
	}

	f.Factory, err = factorypkg.New(
		factorypkg.Option[TAPI]().
			APIConf(f.clientConf.API).
			APIConstructor(f.apiConstructor).
			ConnConf(f.clientConf.Conn).
			InstanceContextKey(f.instanceContextKey).
			Log(f.Log()),
	)
	if err != nil {
		return nil, err
	}

	return f, nil
}
