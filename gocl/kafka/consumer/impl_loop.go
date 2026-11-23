package consumer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/app/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
)

func (s *impl) loopYieldNextProcFn(ctx context.Context, fn YieldNextProcFn, opts runLoopOpts) error {
	ctx = attrname.DictPairsContext(ctx, attrname.Topic, opts.topic)

	reader := s.newReader(ctx, opts.topic)

	defer func() {
		if reader != nil {
			ioext.CloseWithLog(ctx, reader, s.Log())
		}
	}()

	var (
		breakOnFetchNextTimeoutErr   = errors.New("no new message fetched prior the timeout elapsed")
		breakOnLoopRunTimeReachedErr = errors.New("consume loop run time reached")
	)

	if opts.breakOnLoopRunTimeReached != 0 {
		runTimoutCtx, runTimeoutCancel := context.WithTimeoutCause(ctx, opts.breakOnLoopRunTimeReached, breakOnLoopRunTimeReachedErr)

		defer runTimeoutCancel()

		ctx = runTimoutCtx
	}

	var (
		fetchRetriesMax = s.conf.FetchMaxRetries
		noRetriesCtx    = contextext.NoRetry(ctx)
		fetchedMsgCount = 0

		fetchRetriesLeft = fetchRetriesMax
	)

	defer func() {
		s.Log().Trace(ctx).Msgf("totally fetched '%d' messages from topic '%s'", fetchedMsgCount, opts.topic)
	}()

	for {
		var (
			fetchErr error
			msg      kafka.Message
		)

		switch int64(opts.breakOnFetchNextTimeout) {
		case 0:
			msg, fetchErr = reader.FetchMessage(ctx)

			if errors.Is(fetchErr, breakOnLoopRunTimeReachedErr) ||
				(errors.Is(fetchErr, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), breakOnLoopRunTimeReachedErr)) {
				s.Log().
					Warn(ctx).
					Err(fetchErr).
					Msgf("consume loop run time '%s' reached, consume loop exits", opts.breakOnLoopRunTimeReached)

				return nil
			}
		default:
			fetchCtx, fetchCtxCancel := context.WithTimeoutCause(ctx, opts.breakOnFetchNextTimeout, breakOnFetchNextTimeoutErr)

			msg, fetchErr = reader.FetchMessage(fetchCtx)

			if errors.Is(fetchErr, breakOnFetchNextTimeoutErr) ||
				(errors.Is(fetchErr, context.DeadlineExceeded) && errors.Is(context.Cause(fetchCtx), breakOnFetchNextTimeoutErr)) {
				s.Log().
					Warn(fetchCtx).
					Err(fetchErr).
					Msgf("no any message fetched prior the timeout '%s' elapsed, consume loop exits", opts.breakOnFetchNextTimeout)

				fetchCtxCancel()

				return nil
			}

			fetchCtxCancel()
		}

		if fetchErr == nil {
			fetchRetriesLeft = fetchRetriesMax

			fetchedMsgCount++
			if fetchedMsgCount == 1 || fetchedMsgCount%100 == 0 {
				s.Log().Trace(ctx).Msgf("totally fetched '%d' messages from topic '%s'", fetchedMsgCount, opts.topic)
			}

			metricstub.KafkaFetchMsgCount.
				WithLabelVals(opts.topic, s.conf.ConsumerGroupID, strconv.FormatBool(false)).
				Inc()

			msgCtx := attrname.DictPairsContext(
				ctx,
				attrname.Offset, msg.Offset,
				attrname.Partition, msg.Partition,
			)

			next := fn(msgCtx, msg)
			if !next {
				s.Log().
					Warn(msgCtx).
					Msgf("processing of message '%s' reported to break consume loop, consume loop exits", s.msgString(msg))

				return nil
			}

			if opts.noCommits {
				continue
			}

			commitErr := reader.CommitMessages(msgCtx, msg)

			metricstub.KafkaSendCommitMsgCount.
				WithLabelVals(
					msg.Topic,
					s.conf.ConsumerGroupID,
					strconv.FormatBool(commitErr != nil),
				).
				Inc()

			if commitErr == nil {
				s.Log().
					Trace(msgCtx).
					Msgf("succeeded to send commit for message '%s'", s.msgString(msg))

				continue
			}

			commitErr = fmt.Errorf("failed to send commit for message '%s'; %w", s.msgString(msg), commitErr)

			if ctxCause := context.Cause(msgCtx); ctxCause != nil {
				s.Log().
					Err(msgCtx, ctxCause).
					Msgf("error while sending commit for message '%s'; context canceled and consume loop exits", s.msgString(msg))

				return ctxCause
			}

			s.Log().
				Err(ctx, commitErr).
				Msgf("error while sending commit for message '%s'; context canceled and consume loop exits", s.msgString(msg))

			return commitErr
		}

		metricstub.KafkaFetchMsgCount.
			WithLabelVals(opts.topic, s.conf.ConsumerGroupID, strconv.FormatBool(true)).
			Inc()

		if ctxCause := context.Cause(ctx); ctxCause != nil {
			s.Log().
				Err(ctx, ctxCause).
				Msgf("error while fetching next message; context canceled and consume loop exits")

			return ctxCause
		}

		if fetchRetriesLeft == 0 || noRetriesCtx {
			s.Log().
				Err(ctx, fetchErr).
				Msgf("error while fetching next message, no retries left or retries not allowed, cnosume loop exits")

			return fetchErr
		}

		backoffDelay := s.fetchBackoff.Backoff(int(fetchRetriesMax) - int(fetchRetriesLeft))

		fetchRetriesLeft--

		if errors.Is(fetchErr, io.EOF) {
			s.Log().
				Warn(ctx).
				Err(fetchErr).
				Msgf("error while fetching next message, reader unexpectedly closed; new reader will be used"+
					" and next attempt ('%d' from '%d') to fetch will be after backoff '%s'",
					fetchRetriesMax-fetchRetriesLeft,
					fetchRetriesMax,
					backoffDelay)
			ioext.CloseWithLog(ctx, reader, s.Log())

			reader = s.newReader(ctx, opts.topic)
		} else {
			s.Log().
				Warn(ctx).
				Err(fetchErr).
				Msgf("error while fetching next message; next attempt ('%d' from '%d') to fetch will be after backoff '%s'",
					fetchRetriesMax-fetchRetriesLeft,
					fetchRetriesMax,
					backoffDelay)
		}

		select {
		case <-ctx.Done():
			s.Log().
				Err(ctx, context.Cause(ctx)).
				Msgf("error while fetching next message; context canceled and consume loop exits")
		case <-time.After(backoffDelay):
		}
	}
}
