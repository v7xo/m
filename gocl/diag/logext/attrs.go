package logext

import (
	"context"
	"iter"

	attrname "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type attrsLog struct {
	u     Logger
	attrs attrname.Dict
}

func (l *attrsLog) Attrs() iter.Seq2[attrname.Name, any] { return l.attrs.Seq() }
func (l *attrsLog) Debug(ctx context.Context) Event      { return l.u.Trace(ctx).Attrs(l.attrs) }
func (l *attrsLog) Hook() EventHookFn                    { return l.u.Hook() }
func (l *attrsLog) Info(ctx context.Context) Event       { return l.u.Info(ctx).Attrs(l.attrs) }
func (l *attrsLog) IsNop() bool                          { return false }
func (l *attrsLog) LogLevel() level.Num                  { return l.u.LogLevel() }
func (l *attrsLog) Trace(ctx context.Context) Event      { return l.u.Trace(ctx).Attrs(l.attrs) }
func (l *attrsLog) Tune() Tuner                          { return Tune(l) }
func (l *attrsLog) Warn(ctx context.Context) Event       { return l.u.Warn(ctx).Attrs(l.attrs) }

func (l *attrsLog) Err(ctx context.Context, err error) Event { return l.u.Err(ctx, err).Attrs(l.attrs) }

func (l *attrsLog) Fatal(ctx context.Context, err error) Event {
	return l.u.Fatal(ctx, err).Attrs(l.attrs)
}

func (l *attrsLog) Level(ctx context.Context, level level.Num, err error) Event {
	return l.u.Level(ctx, level, err).Attrs(l.attrs)
}

func (l *attrsLog) LevelFromErr(ctx context.Context, err error) (e Event) {
	return l.u.LevelFromErr(ctx, err).Attrs(l.attrs)
}

func Attrs(underlying Logger, attrs attrname.Dict) Logger {
	if underlying.IsNop() {
		return underlying
	}

	return &attrsLog{u: underlying, attrs: attrs}
}
