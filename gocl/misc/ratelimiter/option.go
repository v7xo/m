package ratelimiter

import (
	"fmt"
	"time"

	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/clock"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logging.Logger) OptionFn { return fopt.WithLogScope(o, val, logScope) }

func (o OptionFn) Clock(val clock.Clock) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.clock = val })
}

func (o OptionFn) MinInterval(val time.Duration) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.minEventInterval = val })
}

func (o OptionFn) Jitter(val float64) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.jitter = val })
}

func (o OptionFn) PerMinute(val int64) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		if val < 1 {
			return fmt.Errorf("specified value '%d' of requests per minute is not valid; can't be less than 1", val)
		}

		i.minEventInterval = time.Minute / time.Duration(val)

		return nil
	}
}

func (o OptionFn) PerSecond(val int64) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		if val < 1 {
			return fmt.Errorf("specified value '%d' of requests per second is not valid; can't be less than 1", val)
		}

		i.minEventInterval = time.Second / time.Duration(val)

		return nil
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
