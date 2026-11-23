package retryable

import (
	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/retrier"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/clock"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ratelimiter"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logging.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) Options(val ...OptionFn) OptionFn { return fopt.Pack(o, val...) }

func (o OptionFn) Conf(cfg Conf) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		return applyConf(i, cfg)
	}
}

func (o OptionFn) RetryBufferLimit(val int64) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retrierOpts = i.retrierOpts.RetryBufferLimit(val)

		return
	}
}

func (o OptionFn) ObserveResultFn(val ObserveResultFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.observeResultFn = val

		return
	}
}

func (o OptionFn) RateLimiter(val ratelimiter.Service) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.singlerOpts = i.singlerOpts.RateLimiter(val)

		return
	}
}

// RoundTripFn sets single round-trip function that performs HTTP request using specified transport.
func (o OptionFn) RoundTripFn(val httpext.RoundTripFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.singlerOpts = i.singlerOpts.RoundTripFn(val)

		return
	}
}

// RoundTripTimeout sets single round-trip timeout.
func (o OptionFn) RoundTripTimeout(val timeout.Timeout) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.singlerOpts = i.singlerOpts.RoundTripTimeout(val)

		return
	}
}

func (o OptionFn) NeedRetryFn(val retrier.NeedRetryFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retrierOpts = i.retrierOpts.NeedRetryFn(val)

		return
	}
}

func (o OptionFn) NeedRetryContextFn(val retrier.NeedRetryContextFn) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retrierOpts = i.retrierOpts.NeedRetryContextFn(val)

		return
	}
}

func (o OptionFn) ClientSSL(certPEM, keyPEM []byte) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.singlerOpts = i.singlerOpts.ClientSSL(certPEM, keyPEM)

		return
	}
}

func (o OptionFn) Backoff(val backoff.ExponentialConfig) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retrierOpts = i.retrierOpts.Backoff(val)

		return
	}
}

func (o OptionFn) MaxRetries(val int64) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.retrierOpts = i.retrierOpts.MaxRetries(val)
		i.retryable = val != 0

		return
	}
}

func (o OptionFn) MetricsEnabled(val bool) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.metricsEnabled = val

		return
	}
}

func (o OptionFn) TracingEnabled(val bool) OptionFn {
	return func(i *impl) (err error) {
		if err = o(i); err != nil {
			return
		}

		i.tracingEnabled = val

		return
	}
}

func applyConf(i *impl, cfg Conf) error {
	i.clientName = cfg.ClientName
	i.metricsEnabled = cfg.MetricsEnabled
	i.retryable = cfg.MaxRetries != 0
	i.tracingEnabled = cfg.TracingEnabled

	i.tracingOpts = i.tracingOpts.RequestURLPathFilter(cfg.TracingRequestURLPathFilter)

	i.retrierOpts = i.retrierOpts.Backoff(cfg.Backoff)
	i.retrierOpts = i.retrierOpts.MaxRetries(cfg.MaxRetries)
	i.retrierOpts = i.retrierOpts.RetryBufferLimit(cfg.RetryBufferLimit)

	i.singlerOpts = i.singlerOpts.RoundTripTimeout(cfg.SingleRoundTripTimeout())
	i.singlerOpts = i.singlerOpts.CancelImmediatelyOnRateLimit(cfg.CancelImmediatelyOnRateLimit)
	i.singlerOpts = i.singlerOpts.ClientSSL([]byte(cfg.ClientSSLCertPEM), []byte(cfg.ClientSSLKeyPEM))

	if cfg.DefaultRateLimitPerMinute == 0 {
		i.singlerOpts = i.singlerOpts.RateLimiter(nil)
	} else {
		rl, err := ratelimiter.New(
			ratelimiter.Option().
				Clock(clock.MustDefault(nil)).
				PerMinute(int64(cfg.DefaultRateLimitPerMinute)),
		)
		if err != nil {
			return err
		}

		i.singlerOpts = i.singlerOpts.RateLimiter(rl)
	}

	return nil
}

func Option() OptionFn { return fopt.Nop[impl] }
