package cmdhandler

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/consumer"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ratelimiter"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

// RateLimiter sets rate limiter to use.
//
// See also [OptionFn.RateLimiterOpts].
func (o OptionFn) RateLimiter(val ratelimiter.Service) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.rateLimiter = val })
}

// RateLimiterOpts sets options to be applied to default rate limiter.
//
// Ignored, if [OptionFn.RateLimiter] sets rate limiter.
func (o OptionFn) RateLimiterOpts(val ...ratelimiter.OptionFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.rateLimiterOpts = val })
}

func (o OptionFn) Consumer(val consumer.Service) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.consumer = val })
}

func (o OptionFn) AddHandler(val cmdHandler) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			if val == nil {
				return fmt.Errorf("%w; unable to register nil command handler", errs.ErrNilValue)
			}

			t.cmdHandlers[val.payloadType()] = val

			return nil
		},
	)
}

func Option() OptionFn { return fopt.Nop[impl] }
