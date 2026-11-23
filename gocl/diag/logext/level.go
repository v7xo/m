package logext

import (
	"context"
	"iter"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type levelLog struct {
	u     Logger
	level level.Num
}

var _ Logger = &levelLog{}

func (l *levelLog) Info(ctx context.Context) Event {
	log := l.u

	if l.level < level.InfoNum {
		log = NopLogger
	}

	return log.Info(ctx)
}

func (l *levelLog) Warn(ctx context.Context) Event {
	log := l.u

	if l.level < level.WarnNum {
		log = NopLogger
	}

	return log.Warn(ctx)
}

func (l *levelLog) Err(ctx context.Context, err error) Event {
	log := l.u

	if l.level < level.ErrorNum {
		log = NopLogger
	}

	return log.Err(ctx, err)
}

func (l *levelLog) Fatal(ctx context.Context, err error) Event {
	log := l.u

	if l.level < level.ErrorNum {
		log = NopLogger
	}

	return log.Err(ctx, err)
}

func (l *levelLog) Debug(ctx context.Context) Event {
	log := l.u

	if l.level < level.DebugNum {
		log = NopLogger
	}

	return log.Trace(ctx)
}

func (l *levelLog) Trace(ctx context.Context) Event {
	log := l.u

	if l.level < level.TraceNum {
		log = NopLogger
	}

	return log.Trace(ctx)
}

func (l *levelLog) Level(ctx context.Context, level level.Num, err error) Event {
	log := l.u

	if l.level < level {
		log = NopLogger
	}

	return log.Level(ctx, level, err)
}

func (l *levelLog) LevelFromErr(ctx context.Context, err error) (e Event) {
	return l.Level(ctx, errs.TranslateLevel(err).Num(), err)
}

func (l *levelLog) Attrs() iter.Seq2[attrname.Name, any] { return l.u.Attrs() }
func (l *levelLog) Hook() EventHookFn                    { return l.u.Hook() }
func (l *levelLog) IsNop() bool                          { return false }
func (l *levelLog) LogLevel() level.Num                  { return l.level }
func (l *levelLog) Tune() Tuner                          { return Tune(l) }

func (l *levelLog) WithEventHookFn(hkf EventHookFn) Logger {
	return Hook(l, hkf)
}

func Level(underlying Logger, level level.Num) Logger {
	if !level.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	if underlying.IsNop() || underlying.LogLevel() == level {
		return underlying
	}

	return &levelLog{u: underlying, level: level}
}
