package producer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/app/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logadapter"
	cl_kafka "rn.m.vo7.io/vo7io/m/gocl/kafka"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/header"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type impl struct {
	fopt.LogOpt
	writeBackoff   backoff.Strategy
	closeRequested *atomic.Bool
	dialer         *kafka.Dialer
	writers        *sync.Map
	name           string
	conf           ClientConf
}

func (s *impl) Close() error {
	s.closeRequested.Store(true)

	closeErrs := []error{}

	s.writers.Range(
		func(key, value any) bool {
			w := value.(*kafka.Writer) // nolint

			if closeErr := w.Close(); closeErr != nil {
				closeErrs = append(closeErrs, closeErr)
			}

			return true
		},
	)

	if len(closeErrs) == 0 {
		return nil
	}

	return fmt.Errorf("error while closing producer; %w", errs.Join(closeErrs...))
}

func (s *impl) WriteMessage(ctx context.Context, topic string, key, val []byte, headers map[header.Name]string) error {
	w, err := s.writer(ctx, topic, false)
	if err != nil {
		return err
	}

	msg := s.makeMsg(topic, key, val, headers)

	retriesMax := s.conf.Producer.MaxWriteRetries
	retriesLeft := retriesMax
	noRetry := contextext.NoRetry(ctx)

	for {
		writeErr := w.WriteMessages(ctx, msg)
		if writeErr == nil {
			s.Log().
				Trace(ctx).
				Err(writeErr).
				Msgf("succeeded to write message to topic '%s'", topic)

			metricstub.KafkaProduceMsgCount.
				WithLabelVals(
					s.name,
					topic,
					strconv.FormatBool(false),
				).
				Inc()

			return nil
		}

		metricstub.KafkaProduceMsgCount.
			WithLabelVals(
				s.name,
				topic,
				strconv.FormatBool(true),
			).
			Inc()

		if ctxCause := context.Cause(ctx); ctxCause != nil {
			s.Log().
				Err(ctx, ctxCause).
				Msgf("error while writing message to topic '%s'; context canceled", topic)

			return ctxCause
		}

		if retriesLeft == 0 || noRetry {
			s.Log().
				Err(ctx, writeErr).
				Msgf("error while writing message to topic '%s', no retries left or retries not allowed", topic)

			return writeErr
		}

		backoffDelay := s.writeBackoff.Backoff(int(retriesMax) - int(retriesLeft))

		retriesLeft--

		s.Log().
			Warn(ctx).
			Err(writeErr).
			Msgf("error while writing message to topic '%s'; next attempt ('%d' from '%d') to write will be after backoff '%s'",
				topic,
				retriesMax-retriesLeft,
				retriesMax,
				backoffDelay)

		select {
		case <-ctx.Done():
			s.Log().
				Err(ctx, context.Cause(ctx)).
				Msgf("error while writing message to topic '%s'; context canceled", topic)
		case <-time.After(backoffDelay):
		}
	}
}

func (s *impl) String() string {
	if s == nil {
		return placeholder.Nil
	}

	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString(LogScope)

	sb.WriteString(", name: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(s.name))

	sb.WriteString(", client ID: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(s.conf.Client.ClientID))

	sb.WriteByte(')')

	return sb.String()
}

func (s *impl) makeMsg(topic string, key, val []byte, headers map[header.Name]string) kafka.Message {
	hdrs := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		hdrs = append(hdrs, kafka.Header{Key: k.String(), Value: []byte(v)})
	}

	return kafka.Message{
		Headers: hdrs,
		Key:     key,
		Time:    time.Now(),
		Topic:   topic,
		Value:   val,
	}
}

func (s *impl) writer(ctx context.Context, topic string, renew bool) (*kafka.Writer, error) {
	if topic == "" {
		return nil, errors.New("topic can't be an empty string")
	}

	if s.closeRequested.Load() {
		return nil, fmt.Errorf("%w; producer has beed requested to close", errs.ErrInvalidOperation)
	}

	val, ok := s.writers.Load(topic)
	if ok && !renew {
		return val.(*kafka.Writer), nil // nolint
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(s.conf.Client.Brokers...),
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.Murmur2Balancer{Consistent: true},
		ErrorLogger:            logadapter.PrintfContext(ctx, s.Log(), level.Error),
		Logger:                 logadapter.PrintfContext(ctx, s.Log(), level.Trace),
		MaxAttempts:            1,
		RequiredAcks:           kafka.RequireAll,

		Transport: &kafka.Transport{
			Dial:        s.dialer.DialFunc,
			DialTimeout: s.dialer.Timeout,
			ClientID:    s.dialer.ClientID,
			TLS:         s.dialer.TLS,
		},
	}

	val, loaded := s.writers.LoadOrStore(topic, w)

	switch s.closeRequested.Load() {
	case true:
		if !loaded {
			s.writers.Delete(topic)
		}

		return nil, fmt.Errorf("%w; producer has beed requested to close", errs.ErrInvalidOperation)
	}

	return val.(*kafka.Writer), nil // nolint
}

func (s *impl) initDialer() error {
	dialer, err := cl_kafka.NewDialer(
		cl_kafka.ClientOption().
			Conf(s.conf.Client).
			Log(s.Log()).
			Name(s.name),
	)
	if err != nil {
		return err
	}

	s.dialer = dialer

	return nil
}

func (s *impl) initLog() error {
	return s.AddLogAttrs(
		attrname.Addrs, s.conf.Client.Brokers,
		attrname.ClientID, s.conf.Client.ClientID,
		attrname.NameAttr, placeholder.UnspecifiedIfEmpty(s.name),
	)
}

func New(opts ...OptionFn) (_ Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{writers: &sync.Map{}, closeRequested: &atomic.Bool{}}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if err = impl.conf.Validate(validation.DefaultValidator()); err != nil {
		return nil, err
	}

	impl.writeBackoff = &backoff.Exponential{Conf: impl.conf.Producer.WriteBackoff}

	if err = impl.initLog(); err != nil {
		return nil, err
	}

	if err = impl.initDialer(); err != nil {
		return
	}

	return impl, nil
}
