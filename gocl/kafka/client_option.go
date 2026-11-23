package kafka

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type ClientOptionFn func(*clientBuilder) error

func (o ClientOptionFn) Log(val logext.Logger) ClientOptionFn {
	return fopt.WithLogScope(o, val, ClientLogScope)
}

func (o ClientOptionFn) Name(val string) ClientOptionFn {
	return fopt.Chain(o, func(t *clientBuilder) { t.name = val })
}

func (o ClientOptionFn) TransportContext(val context.Context) ClientOptionFn {
	return fopt.Chain(o, func(t *clientBuilder) { t.transportContext = val })
}

func (o ClientOptionFn) Conf(val ClientConf) ClientOptionFn {
	return fopt.ChainErr(
		o,
		func(t *clientBuilder) error {
			if err := val.Validate(validation.DefaultValidator()); err != nil {
				return err
			}

			t.conf = val

			return nil
		},
	)
}

func ClientOption() ClientOptionFn { return fopt.Nop[clientBuilder] }
