package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/misc/clock"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type impl struct {
	fopt.LogOpt
	clock       clock.Clock
	lastEventTs *atomic.Int64
	fullName    string
	name        string

	minEventInterval time.Duration
	jitter           float64
}

func (s *impl) String() string {
	if s == nil {
		return placeholder.Nil
	}

	var limitString string

	if s.minEventInterval == 0 {
		limitString = "-"
	} else {
		limitString = strconv.FormatFloat(float64(1)/s.minEventInterval.Seconds(), 'f', 6, 64) // nolint:mnd,revive
	}

	return "(" + s.fullName + ", limit: " + limitString + " rps" + ", jitter: " + strconv.FormatFloat(s.jitter, 'f', 3, 64) + ")" // nolint:mnd,revive
}

func (s *impl) WaitNext(ctx context.Context) (time.Time, error) {
	_, ts, err := s.waitNext(ctx, false)

	return ts, err
}

func (s *impl) TryWaitNext(ctx context.Context) (ok bool, ts time.Time, err error) {
	ok, ts, err = s.waitNext(ctx, true)

	return
}

func (s *impl) waitNext(ctx context.Context, try bool) (ok bool, ts time.Time, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while waiting next event on rate limiter '%s'; %w", s.fullName, err)
		}
	}()

	var (
		previouslyReadLastEventTs = missingLastEventTs
		errDelayElapsed           = errors.New("delay elapsed")
	)

	for {
		now := s.clock.Now(ctx).UnixNano()
		lastEventTs := s.lastEventTs.Load()

		if now < 1 {
			err = fmt.Errorf("unexpected condition occurred; current time '%d' (nanosecs) is zero or negative", now)

			return
		} else if previouslyReadLastEventTs > lastEventTs {
			err = fmt.Errorf("unexpected condition occurred; last event timestamp read just now '%s' is less than previously read one '%s'",
				time.Unix(0, lastEventTs),
				time.Unix(0, previouslyReadLastEventTs))

			return
		} else if lastEventTs > now {
			err = fmt.Errorf("unexpected condition occurred; last event timestamp read just now '%s' is greater than current time '%s'",
				time.Unix(0, lastEventTs),
				time.Unix(0, now))

			return
		}

		elapsed := time.Duration(now - lastEventTs)

		if lastEventTs == missingLastEventTs || elapsed >= s.minEventInterval {
			switch {
			case s.lastEventTs.CompareAndSwap(lastEventTs, now):
				ok = true
				ts = time.Unix(0, now)

				s.Log().
					Trace(ctx).
					Msgf("emitting next event; previous event timestamp: %s; current timestamp: %s; min. interval: %s",
						time.Unix(0, lastEventTs),
						ts,
						s.minEventInterval)

				return
			}
		} else if try {
			return
		}

		previouslyReadLastEventTs = lastEventTs
		delay := s.minEventInterval - elapsed

		if s.jitter != 0.0 {
			delay, err = timeout.JitteredSeed(delay, s.jitter, now)
			if err != nil {
				return
			} else if delay == 0 {
				continue
			}
		}

		delayCtx, delayCancel := context.WithTimeoutCause(ctx, delay, errDelayElapsed)

		<-delayCtx.Done()

		err = context.Cause(delayCtx)

		delayCancel()

		if errors.Is(err, errDelayElapsed) {
			err = nil

			continue
		}

		return
	}
}

func New(opts ...OptionFn) (res Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{
		lastEventTs:      &atomic.Int64{},
		jitter:           DefaultJitter,
		name:             DefaultName,
		minEventInterval: DefaultMinInterval,
	}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.jitter < MinJitter {
		return nil, fmt.Errorf("value '%v' of jitter is invalid; can't be less than '%v'", impl.jitter, MinJitter)
	}

	if impl.minEventInterval < time.Duration(1) {
		return nil, fmt.Errorf("value '%s' of min. event interval is invalid; can't be less than '%s'", impl.minEventInterval, time.Duration(1))
	}

	if impl.name == "" {
		return nil, errors.New("no name specified whereas it required")
	}

	impl.fullName = info.Name() + " " + info.CmdName() + "/ " + impl.name

	if impl.clock == nil {
		impl.clock = clock.MustDefault(impl.Log())
	}

	impl.lastEventTs.Swap(missingLastEventTs)

	return impl, nil
}
