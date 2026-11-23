package tracing

import (
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLog(o, val) }

func (o OptionFn) RequestURLPathFilter(val string) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.requestURLPathFilterPattern = val

		return nil
	}
}

func (o OptionFn) Underlying(val http.RoundTripper) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.underlying = val

		return nil
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
