package cmdhandler

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/app/appmodule"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/consumer"
	"rn.m.vo7.io/vo7io/m/gocl/misc/clock"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ratelimiter"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type impl struct {
	fopt.LogOpt

	rateLimiter     ratelimiter.Service
	cmdHandlers     map[reflect.Type]cmdHandler // key is a type of command input payload [CommandInput.Payload].
	consumer        consumer.Service
	rateLimiterOpts []ratelimiter.OptionFn
}

var (
	_ appmodule.Module    = &impl{}
	_ readiness.Collector = &impl{}
)

// Name implements [appmodule.Module.Name].
func (*impl) Name() string { return LogScope }

// Run implements [appmodule.Module.Run].
func (s *impl) Run(ctx context.Context) error {
	return s.consumer.RunLoop(
		ctx,
		consumer.RunLoopOption().
			ProcFn(consumer.YieldNextProcFn(s.procMsgUntilSuccess)),
	)
}

// CollectReadinessConsolidated implements [readiness.Collector.CollectReadinessConsolidated].
func (*impl) CollectReadinessConsolidated(ctx context.Context) (bool, error) {
	// TODO: implement logic evaluating rediness.
	// now just announce OK.
	return true, nil
}

func (*impl) msgString(msg kafka.Message) string {
	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("topic: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(msg.Topic))

	sb.WriteString(", partition: ")
	sb.WriteString(stringsext.FormatInt(msg.Partition))

	sb.WriteString(", offset: ")
	sb.WriteString(stringsext.FormatInt(msg.Offset))

	sb.WriteString(", hi watermark: ")
	sb.WriteString(stringsext.FormatInt(msg.HighWaterMark))

	sb.WriteByte(')')

	return sb.String()
}

func (s *impl) initRateLimiter() error {
	switch s.rateLimiter {
	case nil:
		var err error

		opt := ratelimiter.Option().
			Clock(clock.MustDefault(s.Log())).
			PerMinute(20).
			Log(s.Log())

		s.rateLimiter, err = ratelimiter.New(fopt.Pack(opt, s.rateLimiterOpts...))
		if err != nil {
			return err
		}
	}

	return nil
}

func New(opts ...OptionFn) (_ Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{cmdHandlers: make(map[reflect.Type]cmdHandler)}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.consumer == nil {
		return nil, fmt.Errorf("no service '%s' specified", typestring.TypeStringForExpr(impl.consumer))
	}

	if err = impl.initRateLimiter(); err != nil {
		return
	}

	return impl, nil
}
