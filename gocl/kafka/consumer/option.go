package consumer

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Conf(val Conf) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			if err := val.Validate(validation.DefaultValidator()); err != nil {
				return err
			}

			t.conf = val
			t.fetchBackoff = &backoff.Exponential{Conf: val.FetchBackoff}

			return nil
		},
	)
}

func Option() OptionFn { return fopt.Nop[impl] }
