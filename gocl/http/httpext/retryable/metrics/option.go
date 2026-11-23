package metrics

import (
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Underlying(val http.RoundTripper) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.underlying = val

		return nil
	}
}

func (o OptionFn) ClientName(val string) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.clientName = val

		return nil
	}
}

func (o OptionFn) RetryableRequest(val bool) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.measureAsRetryableRequest = val

		return nil
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
