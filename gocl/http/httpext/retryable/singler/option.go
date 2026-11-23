package singler

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ratelimiter"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn {
	return fopt.WithLog(o, val)
}

func (o OptionFn) RateLimiter(val ratelimiter.Service) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		i.rateLimiter = val

		return nil
	}
}

func (o OptionFn) ClientSSL(certPEM, keyPEM []byte) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.sslCertPEM = certPEM
		i.sslKeyPEM = keyPEM

		return
	}
}

func (o OptionFn) RoundTripFn(val httpext.RoundTripFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.roundTripFn = val

		return
	}
}

func (o OptionFn) RoundTripTimeout(val timeout.Timeout) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.roundTripTimeout = val

		return
	}
}

func (o OptionFn) CancelImmediatelyOnRateLimit(val bool) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.cancelImmediatelyOnRateLimit = val

		return
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
