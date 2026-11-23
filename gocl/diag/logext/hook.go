package logext

import (
	"context"
	"iter"

	attrname "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext/logtypes"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type EventHookFn = logtypes.EventHookFn

type hookLog struct {
	u    Logger
	hook EventHookFn
}

func (l *hookLog) Attrs() iter.Seq2[attrname.Name, any] { return l.u.Attrs() }
func (l *hookLog) Hook() EventHookFn                    { return l.hook }
func (l *hookLog) IsNop() bool                          { return false }
func (l *hookLog) LogLevel() level.Num                  { return l.u.LogLevel() }
func (l *hookLog) Tune() Tuner                          { return Tune(l) }

func (l *hookLog) Info(ctx context.Context) Event {
	return l.hook(ctx, level.InfoNum, nil, false, l.u.Info(ctx))
}

func (l *hookLog) Warn(ctx context.Context) Event {
	return l.hook(ctx, level.WarnNum, nil, false, l.u.Warn(ctx))
}

func (l *hookLog) Err(ctx context.Context, err error) Event {
	return l.hook(ctx, level.ErrorNum, err, false, l.u.Err(ctx, err))
}

func (l *hookLog) Fatal(ctx context.Context, err error) Event {
	return l.hook(ctx, level.ErrorNum, err, true, l.u.Fatal(ctx, err))
}

func (l *hookLog) Debug(ctx context.Context) Event {
	return l.hook(ctx, level.DebugNum, nil, false, l.u.Trace(ctx))
}

func (l *hookLog) Trace(ctx context.Context) Event {
	return l.hook(ctx, level.TraceNum, nil, false, l.u.Trace(ctx))
}

func (l *hookLog) Level(ctx context.Context, level level.Num, err error) Event {
	return l.hook(ctx, level, err, false, l.u.Level(ctx, level, err))
}

func (l *hookLog) LevelFromErr(ctx context.Context, err error) (e Event) {
	e = l.u.LevelFromErr(ctx, err)

	return l.hook(ctx, e.Level(), err, false, e)
}

func Hook(underlying Logger, h EventHookFn) Logger {
	if underlying.IsNop() {
		return underlying
	}

	return &hookLog{u: underlying, hook: h}
}
