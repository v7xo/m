package logtypes

import (
	"context"
	"iter"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type Logger interface {
	// IsNop indicates whether this logger is NoOp - no operational (mock, dummy).
	IsNop() bool
	// LogLevel returns logging level filter for this logger.
	LogLevel() level.Num
	Attrs() iter.Seq2[attrname.Name, any]
	// Tune returns tuner used to create a new logger with overridden properties.
	Tune() Tuner
	Hook() EventHookFn

	// Info creates log event with information level.
	Info(ctx context.Context) Event
	// Warn creates log event with warning level.
	Warn(ctx context.Context) Event
	// Err creates log event with error level.
	Err(ctx context.Context, err error) Event
	// Fatal creates log event with error level.
	// Program will be exited (using [ospkg.Exit] call) immediately after fatal event will delivered to log.
	Fatal(ctx context.Context, err error) Event
	// Trace creates log event with trace level.
	Trace(ctx context.Context) Event
	// Level creates logger event with specified level.
	// Method will panic with in case of unknown level.
	Level(ctx context.Context, l level.Num, err error) Event
	// LevelFromErr same as [Logger.Level] but logging level determined by specified err.
	LevelFromErr(ctx context.Context, err error) Event
}

type Tuner interface {
	AddAttrs(attrname.Dict) Tuner
	Level(l level.Num) Tuner
	NextHook(EventHookFn) Tuner
	SetAttrs(attrname.Dict) Tuner
	SetHook(EventHookFn) Tuner

	// Tuned returns new logger with applied tuning options.
	Tuned() Logger
}

type EventHookFn func(ctx context.Context, l level.Num, err error, fatal bool, e Event) Event

type Event interface {
	Bool(k attrname.Name, v bool) Event
	Str(k attrname.Name, v string) Event
	Float32(k attrname.Name, v float32) Event
	Float64(k attrname.Name, v float64) Event
	Int(k attrname.Name, v int) Event
	Int32(k attrname.Name, v int32) Event
	Int64(k attrname.Name, v int64) Event
	Uint(k attrname.Name, v uint) Event
	Uint32(k attrname.Name, v uint32) Event
	Uint64(k attrname.Name, v uint64) Event
	Attrs(d attrname.Dict) Event
	Err(err error) Event
	// Context returns context associated with this event (see Logger.{Info,Warn,Err,Fatal,Trace} method).
	//
	// If no context associated then returns [context.Background].
	Context() context.Context
	// Level returns logging level of this event.
	Level() level.Num
	// WithSendMsg sets log event message field.
	//
	// Consider, that this message can be overridden by Msg, Msgf call.
	WithSendMsg(msg string) Event
	// WithTimestamp sets log event timestamp field.
	WithTimestamp(ts time.Time) Event
	// Msg sends log event to output with specified message.
	Msg(msg string)
	// Msgf sends log event to output with specified message.
	Msgf(msg string, v ...any)
	// Send sends log event to output.
	Send()
	IsNop() bool
}
