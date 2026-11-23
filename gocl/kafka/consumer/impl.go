package consumer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logadapter"
	cl_kafka "rn.m.vo7.io/vo7io/m/gocl/kafka"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type impl struct {
	fopt.LogOpt
	fetchBackoff backoff.Strategy
	dialer       *kafka.Dialer
	conf         Conf
}

func (s *impl) RunLoop(ctx context.Context, opts ...RunLoopOptionFn) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error occurred during Kafka consumer loop; %w", err)

			s.Log().Err(ctx, err).Send()
		}
	}()

	appliedOpts := runLoopOpts{}
	if err = fopt.Apply(&appliedOpts, opts...); err != nil {
		return err
	}

	if appliedOpts.topic == "" {
		appliedOpts.topic = s.conf.Topic
	}

	if appliedOpts.topic == "" {
		return errors.New("no topic name specified, neither by consumer configuration nor by loop options")
	}

	switch procFn := appliedOpts.procFn.(type) {
	case YieldNextProcFn:
		return s.loopYieldNextProcFn(ctx, procFn, appliedOpts)
	case nil:
		return errors.New("no message processing function specified")
	default:
		return fmt.Errorf("message processing function '%s' not supported", stringsext.TypeStringOfVal(appliedOpts.procFn))
	}
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

func (s *impl) newReader(ctx context.Context, topic string) *kafka.Reader {
	readerConf := kafka.ReaderConfig{
		Brokers:        s.conf.Client.Brokers,
		CommitInterval: 0,
		Dialer:         s.dialer,
		ErrorLogger:    logadapter.PrintfContext(ctx, s.Log(), level.Error),
		GroupBalancers: []kafka.GroupBalancer{kafka.RangeGroupBalancer{}},
		GroupID:        s.conf.ConsumerGroupID,
		Topic:          topic,
	}

	if s.conf.EmitReaderInfoLog {
		readerConf.Logger = logadapter.PrintfContext(ctx, s.Log(), level.Info)
	}

	return kafka.NewReader(readerConf)
}

func (s *impl) initDialer() error {
	dialer, err := cl_kafka.NewDialer(cl_kafka.ClientOption().Conf(s.conf.Client).Log(s.Log()))
	if err != nil {
		return err
	}

	s.dialer = dialer

	return nil
}

func New(opts ...OptionFn) (_ Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.conf.FetchMaxRetries < 0 {
		return nil, fmt.Errorf("invalid value '%d' of max. fetch retries, can't be less than 0", impl.conf.FetchMaxRetries)
	}

	if impl.fetchBackoff == nil && impl.conf.FetchMaxRetries != 0 {
		return nil, errors.New("no backoff strategy set for fetch retries")
	}

	err = impl.AddLogAttrs(
		attrname.Addrs, impl.conf.Client.Brokers,
		attrname.ConsumerGroup, impl.conf.ConsumerGroupID,
		attrname.ClientID, impl.conf.Client.ClientID,
	)
	if err != nil {
		return nil, err
	}

	if err = impl.initDialer(); err != nil {
		return nil, err
	}

	return impl, nil
}
