package clock

import (
	"context"
	"fmt"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type impl struct {
	fopt.LogOpt
	nowFn func(context.Context) time.Time
}

func (s *impl) CurrentTimestamp(ctx context.Context) time.Time { return s.Now(ctx) }

func (s *impl) Now(ctx context.Context) time.Time { return s.nowFn(ctx) }

func (*impl) defaultNow(_ context.Context) time.Time { return time.Now() }

func New(opts ...OptionFn) (_ Clock, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create clock service; %w", err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.nowFn == nil {
		impl.nowFn = impl.defaultNow
	}

	return impl, nil
}

func MustDefault(log logext.Logger) Clock {
	res, err := New(Option().Log(log))
	if err != nil {
		panic(err)
	}

	return res
}
