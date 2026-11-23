package gokit

//nolint:gci
import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"os"

	kit_log "github.com/go-kit/log"
	kit_level "github.com/go-kit/log/level"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type impl struct {
	nonErrLevelWriter io.Writer
	errLevelWriter    io.Writer

	kitDebugLevelLog kit_log.Logger
	kitErrLevelLog   kit_log.Logger
	kitInfoLevelLog  kit_log.Logger
	kitWarnLevelLog  kit_log.Logger

	callerInfo callerInfoFn

	infoEventConstructor,
	debugEventConstructor,
	warnEventConstructor func(ctx context.Context) logext.Event

	includeAttrFn, excludeAttrFn func(attrname.Name) bool

	// log event (record) additional attributes.
	appCmdName func() string
	appName    func() string
	appVersion func() string

	format logext.Format
	level  level.Num
}

func (*impl) Attrs() iter.Seq2[attrname.Name, any]       { return maps.All(attrname.Dict(nil)) }
func (*impl) Hook() logext.EventHookFn                   { return nil }
func (*impl) IsNop() bool                                { return false }
func (l *impl) Debug(ctx context.Context) logext.Event   { return l.debugEventConstructor(ctx) }
func (l *impl) Info(ctx context.Context) logext.Event    { return l.infoEventConstructor(ctx) }
func (l *impl) LogLevel() level.Num                      { return l.level }
func (l *impl) newNopEvent(context.Context) logext.Event { return &event{nop: true, l: l.level} }
func (l *impl) Trace(ctx context.Context) logext.Event   { return l.debugEventConstructor(ctx) }
func (l *impl) Tune() logext.Tuner                       { return logext.Tune(l) }
func (l *impl) Warn(ctx context.Context) logext.Event    { return l.warnEventConstructor(ctx) }

func (l *impl) Err(ctx context.Context, err error) logext.Event {
	return newEventErr(l, ctx, l.kitErrLevelLog, err)
}

func (l *impl) Fatal(ctx context.Context, err error) logext.Event {
	return newEventFatal(l, ctx, l.kitErrLevelLog, err)
}

func (s *impl) Level(ctx context.Context, l level.Num, err error) (e logext.Event) {
	switch {
	case l.IsInfo():
		e = s.infoEventConstructor(ctx)
	case l.IsWarn():
		e = s.warnEventConstructor(ctx)
	case l.IsError():
		e = s.Err(ctx, err)

		return
	case l.IsTraceOrDebug():
		e = s.debugEventConstructor(ctx)
	default:
		panic(errs.ErrLogUnknownLevel)
	}

	if err != nil {
		e = e.Err(err)
	}

	return
}

func (l *impl) LevelFromErr(ctx context.Context, err error) (e logext.Event) {
	return l.Level(ctx, errs.TranslateLevel(err).Num(), err)
}

func (l *impl) newDebugLevelLogEvent(ctx context.Context) logext.Event {
	return newEvent(l, ctx, l.kitDebugLevelLog, level.DebugNum)
}

func (l *impl) newWarnLevelLogEvent(ctx context.Context) logext.Event {
	return newEvent(l, ctx, l.kitWarnLevelLog, level.WarnNum)
}

func (l *impl) newInfoLevelLogEvent(ctx context.Context) logext.Event {
	return newEvent(l, ctx, l.kitInfoLevelLog, level.InfoNum)
}

//nolint:maintidx
func (l *impl) initLevelLoggers(nonErrLevel kit_log.Logger, errLevel kit_log.Logger) {
	// error & fatal - always logged
	l.kitErrLevelLog = kit_level.Error(errLevel)

	// debug/trace, info, warn - skips log by default
	l.debugEventConstructor = l.newNopEvent
	l.warnEventConstructor = l.newNopEvent
	l.infoEventConstructor = l.newNopEvent

	levelNum := l.level

	if level.DebugNum <= levelNum {
		l.kitDebugLevelLog = kit_level.Debug(nonErrLevel)
		l.debugEventConstructor = l.newDebugLevelLogEvent
	}

	if level.WarnNum <= levelNum {
		l.kitWarnLevelLog = kit_level.Warn(nonErrLevel)
		l.warnEventConstructor = l.newWarnLevelLogEvent
	}

	if level.InfoNum <= levelNum {
		l.kitInfoLevelLog = kit_level.Info(nonErrLevel)
		l.infoEventConstructor = l.newInfoLevelLogEvent
	}
}

func (l *impl) emitAttr(n attrname.Name) bool {
	if l.includeAttrFn == nil && l.excludeAttrFn == nil {
		return true
	} else if l.includeAttrFn == nil {
		return !l.excludeAttrFn(n)
	}

	return l.includeAttrFn(n)
}

func (l *impl) initNewGoKitLogger(writer io.Writer) (kit_log.Logger, error) {
	if writer == nil {
		return nil, errors.New("no logger writer specified (nil)")
	}

	var logger kit_log.Logger

	switch {
	case l.format.IsJSON():
		logger = kit_log.NewJSONLogger(writer)
	case l.format.IsLogfmt():
		logger = kit_log.NewLogfmtLogger(writer)
	default:
		return nil, fmt.Errorf("logger output format '%s' is unknown", l.format)
	}

	logger = kit_level.NewFilter(logger, kit_level.AllowAll())

	if l.appName != nil && l.emitAttr(attrname.AppName) {
		logger = kit_log.With(logger, string(attrname.AppName), kit_log.Valuer(func() any { return placeholder.UnspecifiedIfEmpty(l.appName()) }))
	}

	if l.appVersion != nil && l.emitAttr(attrname.AppVersion) {
		logger = kit_log.With(logger, string(attrname.AppVersion), kit_log.Valuer(func() any { return placeholder.UnspecifiedIfEmpty(l.appVersion()) }))
	}

	if l.appCmdName != nil && l.emitAttr(attrname.AppCmdName) {
		logger = kit_log.With(logger, string(attrname.AppCmdName), kit_log.Valuer(func() any { return placeholder.UnspecifiedIfEmpty(l.appCmdName()) }))
	}

	return logger, nil
}

func (l *impl) init() error {
	if l.nonErrLevelWriter == nil {
		l.nonErrLevelWriter = kit_log.NewSyncWriter(os.Stderr)
	}

	if l.errLevelWriter == nil {
		l.errLevelWriter = kit_log.NewSyncWriter(os.Stderr)
	}

	nonErrLevelLog, err := l.initNewGoKitLogger(l.nonErrLevelWriter)
	if err != nil {
		return err
	}

	errLevelLog, err := l.initNewGoKitLogger(l.errLevelWriter)
	if err != nil {
		return err
	}

	l.initLevelLoggers(nonErrLevelLog, errLevelLog)

	return nil
}

func New(opts ...OptionFn) (res logext.Logger, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while creating instance of Go-kit logger; err: %w", err)
		}
	}()

	impl := &impl{}

	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if err = impl.init(); err != nil {
		return
	}

	return impl, nil
}
