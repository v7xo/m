package retrier

import (
	"context"
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn {
	return fopt.WithLog(o, val)
}

func (o OptionFn) RetryBufferLimit(val int64) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retryBufLimit = val

		return
	}
}

func (o OptionFn) NeedRetryFn(val NeedRetryFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		if val == nil {
			i.needRetryFn = nil
		}

		i.needRetryFn = func(_ context.Context, r *http.Response) (bool, error) {
			return val(r)
		}

		return
	}
}

func (o OptionFn) NeedRetryContextFn(val NeedRetryContextFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.needRetryFn = val

		return
	}
}

func (o OptionFn) Backoff(val backoff.ExponentialConfig) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.backoff = &backoff.Exponential{Conf: val}

		return
	}
}

func (o OptionFn) MaxRetries(val int64) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.maxRetries = val

		return
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
