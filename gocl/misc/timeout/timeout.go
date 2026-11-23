package timeout

import (
	"context"
	"fmt"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Timeout time.Duration

const Infinite Timeout = -1

func (t Timeout) IsInfinite() bool {
	return t == Infinite
}

func (t Timeout) IsZero() bool       { return t == 0 }
func (t Timeout) Dur() time.Duration { return time.Duration(t) }
func (t Timeout) IsValid() bool      { return t.IsInfinite() || t >= 0 }

func (t Timeout) String() string {
	if t.IsInfinite() {
		return placeholder.Infinity
	}

	return time.Duration(t).String()
}

// Milliseconds returns the timeout as an integer millisecond count or [Infinite].
func (t Timeout) Milliseconds() int64 {
	if t == Infinite {
		return int64(Infinite)
	}

	return time.Duration(t).Milliseconds()
}

func (t Timeout) Context(p context.Context) (context.Context, context.CancelFunc) {
	return t.ContextCause(p, nil)
}

func (t Timeout) ContextCause(p context.Context, cause error) (context.Context, context.CancelFunc) {
	if t == Infinite {
		return p, func() {}
	}

	cause = errs.Join(cause, fmt.Errorf("timeout '%s' elapsed; %w", t.String(), context.DeadlineExceeded))

	return context.WithTimeoutCause(p, t.Dur(), cause)
}

// FromMs returns Timeout converted from an integer millesecond count.
func FromMs(ms int) Timeout {
	if ms == int(Infinite) {
		return Infinite
	}

	return Timeout(time.Millisecond * time.Duration(ms))
}
