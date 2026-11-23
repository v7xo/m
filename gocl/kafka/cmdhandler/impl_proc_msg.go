package cmdhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/header"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func (s *impl) procMsgUntilSuccess(ctx context.Context, msg kafka.Message) (yieldNext bool) {
	for {
		if next, retry := s.procMsg(ctx, msg); next {
			return true
		} else if retry {
			continue
		}

		return false
	}
}

func (s *impl) procMsg(ctx context.Context, msg kafka.Message) (yieldNext bool, retry bool) {
	ctx = attrname.DictContext(ctx, s.procMsgLogAttrs(uuidext.MustRand4BHex()))

	defer func() {
		if panicErr := errs.PanicErr(recover()); panicErr != nil {
			s.Log().Err(ctx, panicErr).Msg("panic occurred while processing message")

			yieldNext, retry = false, false
		}
	}()

	_, err := s.rateLimiter.WaitNext(ctx)
	if err != nil {
		s.Log().Err(ctx, err).Msg("error occurred while waiting on message process rate limiter")

		return false, false
	}

	hdrs, ok := s.procMsgReadMsgHeaders(ctx, msg)
	if !ok {
		s.Log().
			Warn(ctx).
			Msgf("reading headers of message wasn't succeesful, skip processing for message '%s' and move to next message", s.msgString(msg))

		return true, false
	}

	ctx = context.WithValue(ctx, contextext.KeyCorrelationID, hdrs[header.NameCorrelationID])
	ctx = context.WithValue(ctx, contextext.KeyTraceParent, hdrs[header.NameTraceParent])
	ctx = attrname.DictContext(ctx, hdrs.ToAttrDict())

	cmdInp, ok := s.procMsgReadCmdInput(ctx, msg, hdrs)
	if !ok {
		s.Log().
			Err(ctx, nil).
			Msgf("reading command input wasn't succeesful for message '%s', unable to move to next message", s.msgString(msg))

		return false, false
	}

	ctx = attrname.DictContext(ctx, cmdInp.ToAttrDict())

	if !s.procMsgHandleCmd(ctx, cmdInp) {
		s.Log().
			Err(ctx, nil).
			Msgf("command '%s' for message '%s' wasn't succeesfully handled, unable to move to next message", cmdInp.String(), s.msgString(msg))

		return false, true
	}

	s.Log().
		Info(ctx).
		Msgf("succeeded to handle command '%s' for message '%s', move to next message", cmdInp.String(), s.msgString(msg))

	return true, false
}

func (s *impl) procMsgReadMsgHeaders(ctx context.Context, msg kafka.Message) (header.Coll, bool) {
	var err error

	defer func() {
		if err != nil {
			s.Log().Err(ctx, err).Msgf("failed to read headers of message '%s'", s.msgString(msg))

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to read headers of message '%s'", s.msgString(msg))
	}()

	res := make(header.Coll)

	err = res.ReadFromKafkaMessage(
		msg,
		header.CodeCorrelationID|header.CodeTraceParent,
		header.CodeCorrelationID,
	)
	if err != nil {
		return nil, false
	}

	return res, true
}

func (s *impl) procMsgReadCmdInput(ctx context.Context, msg kafka.Message, hdrs header.Coll) (CmdInput, bool) {
	var err error

	defer func() {
		if err != nil {
			s.Log().Err(ctx, err).Msgf("failed to read command input from message '%s'", s.msgString(msg))

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to read command input from message '%s'", s.msgString(msg))
	}()

	var res CmdInput

	if err = json.Unmarshal(msg.Value, &res); err != nil {
		return CmdInput{}, false
	}

	res.CorrelationID = hdrs[header.NameCorrelationID]
	res.TraceParent = hdrs[header.NameTraceParent]

	return res, true
}

func (s *impl) procMsgHandleCmd(ctx context.Context, cmd CmdInput) bool {
	var err error

	defer func() {
		if err != nil {
			s.Log().Err(ctx, err).Msgf("failed to handle command '%s'", cmd)
		}
	}()

	if err = cmd.Payload.Val.Validate(); err != nil {
		return false
	}

	payloadTyp := reflect.TypeOf(cmd.Payload.Val)
	switch {
	case payloadTyp == nil:
		err = fmt.Errorf("%w; nil payload", errs.ErrNilValue)
	default:
		if hnd, ok := s.cmdHandlers[payloadTyp]; ok {
			err = hnd.invoke(ctx, cmd)

			break
		}

		err = fmt.Errorf("unknown type '%s' of command payload", stringsext.TypeStringOfVal(cmd.Payload.Val))
	}

	return err == nil
}

func (*impl) procMsgLogAttrs(procRunID string) attrname.Dict {
	return attrname.Dict{attrname.MessageProcessingRunID: procRunID}
}
