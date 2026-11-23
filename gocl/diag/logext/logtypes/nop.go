package logtypes

import (
	"context"
	"iter"
	"maps"
	"os"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type nopEvent struct {
	level level.Num
	fatal bool
}

//nolint:gochecknoglobals
var (
	trcNopEvent   Event = &nopEvent{level: level.TraceNum}
	dbgNopEvent   Event = &nopEvent{level: level.DebugNum}
	infNopEvent   Event = &nopEvent{level: level.InfoNum}
	wrnNopEvent   Event = &nopEvent{level: level.WarnNum}
	errNopEvent   Event = &nopEvent{level: level.ErrorNum}
	fatalNopEvent Event = &nopEvent{level: level.ErrorNum, fatal: true}
)

func (*nopEvent) Context() context.Context                   { return context.Background() }
func (*nopEvent) IsNop() bool                                { return true }
func (e *nopEvent) Attrs(f attrname.Dict) Event              { return e }
func (e *nopEvent) Bool(k attrname.Name, v bool) Event       { return e }
func (e *nopEvent) Err(err error) Event                      { return e }
func (e *nopEvent) Float32(k attrname.Name, v float32) Event { return e }
func (e *nopEvent) Float64(k attrname.Name, v float64) Event { return e }
func (e *nopEvent) Int(k attrname.Name, v int) Event         { return e }
func (e *nopEvent) Int32(k attrname.Name, v int32) Event     { return e }
func (e *nopEvent) Int64(k attrname.Name, v int64) Event     { return e }
func (e *nopEvent) Level() level.Num                         { return e.level }
func (e *nopEvent) Str(k attrname.Name, v string) Event      { return e }
func (e *nopEvent) Uint(k attrname.Name, v uint) Event       { return e }
func (e *nopEvent) Uint32(k attrname.Name, v uint32) Event   { return e }
func (e *nopEvent) Uint64(k attrname.Name, v uint64) Event   { return e }
func (e *nopEvent) WithSendMsg(msg string) Event             { return e }
func (e *nopEvent) WithTimestamp(ts time.Time) Event         { return e }

func (e *nopEvent) Send() {
	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

func (e *nopEvent) Msg(msg string) {
	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

func (e *nopEvent) Msgf(msg string, v ...any) {
	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

type nopLog struct{}

//nolint:gochecknoglobals
var NopLogger Logger = &nopLog{}

func (*nopLog) Debug(ctx context.Context) Event            { return dbgNopEvent }
func (*nopLog) Err(ctx context.Context, err error) Event   { return errNopEvent }
func (*nopLog) Fatal(ctx context.Context, err error) Event { return fatalNopEvent }
func (*nopLog) Hook() EventHookFn                          { return nil }
func (*nopLog) Info(ctx context.Context) Event             { return infNopEvent }
func (*nopLog) IsNop() bool                                { return true }
func (*nopLog) LogLevel() level.Num                        { return level.ErrorNum }
func (*nopLog) Trace(ctx context.Context) Event            { return trcNopEvent }
func (*nopLog) Warn(ctx context.Context) Event             { return wrnNopEvent }
func (l *nopLog) Attrs() iter.Seq2[attrname.Name, any]     { return maps.All(attrname.Dict(nil)) }
func (l *nopLog) Tune() Tuner                              { return &nopTuner{base: l} }

func (*nopLog) Level(ctx context.Context, l level.Num, err error) Event {
	switch {
	case l.IsInfo():
		return infNopEvent
	case l.IsWarn():
		return wrnNopEvent
	case l.IsError():
		return errNopEvent
	case l.IsDebug():
		return dbgNopEvent
	case l.IsTrace():
		return trcNopEvent
	default:
		panic(errs.ErrLogUnknownLevel)
	}
}

func (l *nopLog) LevelFromErr(ctx context.Context, err error) Event {
	return l.Level(ctx, errs.TranslateLevel(err).Num(), err)
}

type nopTuner struct{ base Logger }

var _ Tuner = &nopTuner{}

func (t *nopTuner) AddAttrs(a attrname.Dict) Tuner { return t }
func (t *nopTuner) NextHook(h EventHookFn) Tuner   { return t }
func (t *nopTuner) SetAttrs(a attrname.Dict) Tuner { return t }
func (t *nopTuner) SetHook(h EventHookFn) Tuner    { return t }
func (t *nopTuner) Tuned() Logger                  { return t.base }

func (t *nopTuner) Level(l level.Num) Tuner {
	if !l.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	return t
}
