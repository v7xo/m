package gokit

//nolint:gci
import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	kit_log "github.com/go-kit/log"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type (
	eventAttr struct {
		v any
		n attrname.Name
	}

	eventAttrList []eventAttr

	event struct {
		kitLog       kit_log.Logger
		ctx          context.Context //nolint:containedctx
		err          error
		log          *impl
		tsFn         func() time.Time
		tsUnixNanoFn func() int64
		sendMsg      string
		attrs        eventAttrList
		l            level.Num
		fatal        bool
		nop          bool
	}
)

var (
	eventAttrFilterMaps *sync.Pool = &sync.Pool{} //nolint:gochecknoglobals
	eventTailAttrNames             = map[string]struct{}{
		string(attrname.Message): {},
		string(attrname.Err):     {},
	}
)

func (winner eventAttrList) mergeKitLogPairsTo(emit func(attrname.Name) bool, loser attrname.Dict, to []any) []any {
	if len(winner) == 0 && len(loser) == 0 {
		return to
	}

	var m map[string]struct{}

	if mPooled, ok := eventAttrFilterMaps.Get().(map[string]struct{}); ok {
		m = mPooled
	} else {
		m = make(map[string]struct{}, eventAttrsCap)
	}

	tail := make([]any, 0, len(eventTailAttrNames)*2) //nolint:mnd,revive

	for i := len(winner) - 1; i > -1; i-- {
		na := winner[i].n
		ns := string(na)
		v := winner[i].v

		if _, ok := m[ns]; ok {
			continue
		}

		m[ns] = struct{}{}

		if !emit(na) {
			continue
		}

		if _, ok := eventTailAttrNames[ns]; ok {
			tail = append(tail, ns, v)

			continue
		}

		to = append(to, ns, v)
	}

	for na, v := range loser {
		ns := string(na)

		if _, ok := m[ns]; ok {
			continue
		}

		if !emit(na) {
			continue
		}

		if _, ok := eventTailAttrNames[ns]; ok {
			tail = append(tail, ns, v)

			continue
		}

		to = append(to, ns, v)
	}

	for k := range m {
		delete(m, k)
	}

	eventAttrFilterMaps.Put(m)

	to = append(to, tail...)

	return to
}

func (e *event) Bool(k attrname.Name, v bool) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Str(k attrname.Name, v string) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Float32(k attrname.Name, v float32) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Float64(k attrname.Name, v float64) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Int(k attrname.Name, v int) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Int32(k attrname.Name, v int32) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Int64(k attrname.Name, v int64) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Uint(k attrname.Name, v uint) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Uint32(k attrname.Name, v uint32) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Uint64(k attrname.Name, v uint64) logging.Event {
	if !e.nop {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) Attrs(attrs attrname.Dict) logging.Event {
	if e.nop || len(attrs) == 0 {
		return e
	}

	for k, v := range attrs {
		e.attrs = append(e.attrs, eventAttr{n: k, v: v})
	}

	return e
}

func (e *event) WithSendMsg(msg string) logging.Event {
	if !e.nop {
		e.sendMsg = msg
	}

	return e
}

func (e *event) WithTimestamp(ts time.Time) logging.Event {
	if !e.nop {
		e.tsFn = func() time.Time { return ts }

		tsUTC := ts.UTC()

		e.tsUnixNanoFn = func() int64 { return tsUTC.UnixNano() }
	}

	return e
}

func (e *event) Err(err error) logging.Event {
	if !e.nop {
		e.err = err
	}

	return e
}

func (e *event) IsNop() bool {
	return e.nop
}

func (e *event) Context() context.Context {
	if e.ctx == nil {
		return context.Background()
	}

	return e.ctx
}

func (e *event) Send() {
	if e.nop {
		return
	}

	_ = e.kitLog.Log(e.formKitLogPairs(e.sendMsg, false)...) // TODO: handle error

	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

func (e *event) Msg(msg string) {
	if e.nop {
		return
	}

	_ = e.kitLog.Log(e.formKitLogPairs(msg, false)...) // TODO: handle error

	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

func (e *event) Msgf(msg string, v ...any) {
	if e.nop {
		return
	}

	_ = e.kitLog.Log(e.formKitLogPairs(msg, true, v...)...) // TODO: handle error

	if e.fatal {
		os.Exit(errs.FailExitCode)
	}
}

func (e *event) Level() level.Num {
	return e.l
}

// nolint:gocognit,maintidx
func (e *event) formKitLogPairs(msg string, msgf bool, msgfArgs ...any) []any {
	tsStr := e.tsFn().UTC().Format(logging.DefaultTimestampFieldFormat)
	tsUnixNano := e.tsUnixNanoFn()
	wonAttrs := e.attrs
	loserAttrs, _, _ := contextext.AttrDict(e.ctx) // TODO: handle error.

	if e.log.callerInfo != nil {
		fnName, fileLine := e.log.callerInfo()

		wonAttrs = append(
			wonAttrs,
			eventAttr{n: attrname.File, v: fileLine},
			eventAttr{n: attrname.Func, v: fnName},
		)
	}

	if len(msg) != 0 {
		if msgf {
			wonAttrs = append(wonAttrs, eventAttr{n: attrname.Message, v: fmt.Sprintf(msg, msgfArgs...)})
		} else {
			wonAttrs = append(wonAttrs, eventAttr{n: attrname.Message, v: msg})
		}
	}

	if e.err != nil {
		wonAttrs = append(wonAttrs, eventAttr{n: attrname.Err, v: e.err.Error()})
	}

	if e.ctx != nil {
		if goTestName, ok := contextext.TestName(e.ctx); ok && len(goTestName) != 0 {
			wonAttrs = append(wonAttrs, eventAttr{n: attrname.GoTestName, v: goTestName})
		}

		if traceID := contextext.TraceIDString(e.ctx); len(traceID) != 0 {
			wonAttrs = append(wonAttrs, eventAttr{n: attrname.TraceID, v: traceID})
		}

		if spanID := contextext.SpanIDString(e.ctx); len(spanID) != 0 {
			wonAttrs = append(wonAttrs, eventAttr{n: attrname.SpanID, v: spanID})
		}
	}

	wonAttrs = append(wonAttrs, eventAttr{n: attrname.TimestampUnixNano, v: tsUnixNano})
	wonAttrs = append(wonAttrs, eventAttr{n: attrname.DateTime, v: tsStr})

	return wonAttrs.mergeKitLogPairsTo(e.log.emitAttr, loserAttrs, make([]any, 0, (len(wonAttrs)+len(loserAttrs))*2)) //nolint:mnd
}

func newEvent(log *impl, ctx context.Context, kitLog kit_log.Logger, l level.Num) logging.Event {
	return &event{
		ctx:          ctx,
		attrs:        make(eventAttrList, 0, eventAttrsCap),
		kitLog:       kitLog,
		log:          log,
		tsFn:         time.Now,
		tsUnixNanoFn: tsUnixNano,
		l:            l,
	}
}

func newEventFatal(log *impl, ctx context.Context, kitLog kit_log.Logger, err error) logging.Event {
	return &event{
		ctx:          ctx,
		err:          err,
		fatal:        true,
		attrs:        make(eventAttrList, 0, eventAttrsCap),
		kitLog:       kitLog,
		log:          log,
		tsFn:         time.Now,
		tsUnixNanoFn: tsUnixNano,
		l:            level.ErrorNum,
	}
}

func newEventErr(log *impl, ctx context.Context, kitLog kit_log.Logger, err error) logging.Event {
	return &event{
		ctx:          ctx,
		err:          err,
		attrs:        make(eventAttrList, 0, eventAttrsCap),
		kitLog:       kitLog,
		log:          log,
		tsFn:         time.Now,
		tsUnixNanoFn: tsUnixNano,
		l:            level.ErrorNum,
	}
}

func tsUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}
