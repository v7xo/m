package retrier

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
	backoff       backoff.Strategy
	needRetryFn   NeedRetryContextFn
	underlying    http.RoundTripper
	maxRetries    int64
	retryBufLimit int64
}

// RoundTrip implements [http.RoundTripper] interface.
func (s *impl) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req, logAttrs, err := httpext.RequestWithAttrs(req, false)
	if err != nil {
		return
	}

	maxRetries := s.maxRetries

	if contextext.NoRetry(req.Context()) {
		maxRetries = 0
	}

	logAttrs[attrname.RetryCount] = maxRetries
	req = req.WithContext(attrname.DictContext(req.Context(), logAttrs.ToDict()))

	ctx := req.Context()

	if maxRetries != 0 {
		s.Log().Trace(ctx).Msgf("request '%s %s', max. retries '%d'", req.Method, req.URL.Redacted(), maxRetries)
	}

	var reqBodyBuf *[]byte

	if req.Body != nil && maxRetries > 0 {
		retryBufWriter := bytes.NewBuffer(nil)

		_, err = ioext.CopyMaxN(retryBufWriter, req.Body, s.retryBufLimit)
		if err != nil {
			err = fmt.Errorf("failed to copy request body to retry buffer; err: %w", err)

			return
		}

		if err = req.Body.Close(); err != nil {
			err = fmt.Errorf("failed to close request body; err: %w", err)

			return
		}

		reqBodyBuf = new(retryBufWriter.Bytes())
		req.Body = io.NopCloser(bytes.NewBuffer(*reqBodyBuf))
	}

	resp, err = s.underlying.RoundTrip(req)
	if err == nil {
		switch {
		case s.needRetryFn != nil:
			needRetry, needRetryErr := s.needRetryFn(ctx, resp)
			if needRetryErr != nil {
				err = fmt.Errorf("function evaluating need of retry resulted with error; err: %w", needRetryErr)

				return
			} else if needRetry {
				resp, err = s.retry(ctx, req, errors.New("request needs to be retried"), reqBodyBuf, maxRetries)

				return
			}
		}

		return
	}

	if errs.RequestRetryable(err) {
		resp, err = s.retry(ctx, req, fmt.Errorf("retryable error occurred; err: %w", err), reqBodyBuf, maxRetries)
	}

	return
}

func (s *impl) retry(ctx context.Context, req *http.Request, reqErr error, reqBodyBuf *[]byte, maxRetries int64) (resp *http.Response, err error) {
	if reqErr == nil {
		err = errors.New("arg 'reqErr' can't be nil")

		return
	}

	attemptedRetries := int64(0)
	logAttrs := make(attrname.Dict, 2) //nolint

retryLoop:
	for attemptedRetries < maxRetries {
		backoffDelay := s.backoff.Backoff(int(attemptedRetries))

		logAttrs[attrname.RetryNum] = attemptedRetries + 1
		logAttrs[attrname.BackoffDelayMs] = backoffDelay.Milliseconds()

		s.Log().
			Warn(ctx).
			Err(reqErr).
			Attrs(logAttrs).
			Msg("request failed; attempt to retry request after backoff delay")

		attemptedRetries++

		if reqErr = delay(ctx, backoffDelay); reqErr != nil {
			reqErr = fmt.Errorf("request retry was not attempted due to context canceled before; err: %w", reqErr)

			break retryLoop
		}

		if reqBodyBuf != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(*reqBodyBuf))
		}

		resp, reqErr = nil, nil // nolint

		resp, reqErr = s.underlying.RoundTrip(req)
		if reqErr == nil {
			switch {
			case s.needRetryFn != nil:
				needRetry, needRetryErr := s.needRetryFn(ctx, resp)
				if needRetryErr != nil {
					reqErr = fmt.Errorf("function evaluating need of retry resulted with error; err: %w", needRetryErr)

					break retryLoop
				} else if needRetry {
					reqErr = errors.New("request needs to be retried")

					continue retryLoop
				}
			}
		} else if errs.RequestRetryable(reqErr) {
			reqErr = fmt.Errorf("retryable error occurred; err: %w", reqErr)

			continue retryLoop
		}

		break retryLoop
	}

	if attemptedRetries > 0 {
		metricData := metricstub.HTTPRequestSecondsDataFromContext(ctx)
		if metricData != nil {
			metricData.Retried = true
		}

		var (
			logEvt logext.Event
			logMsg string
		)

		if reqErr == nil {
			logEvt = s.Log().Info(ctx)
			logMsg = "request retry succeeded (" +
				strconv.FormatInt(attemptedRetries, 10) + " from " + strconv.FormatInt(maxRetries, 10) +
				" retry(-ies) taken)"
		} else {
			logEvt = s.Log().Err(ctx, reqErr)
			logMsg = "request retry failed; non-retryable error occurred or retry limit reached (" +
				strconv.FormatInt(attemptedRetries, 10) + " from " + strconv.FormatInt(maxRetries, 10) +
				" retry(-ies) taken)"
		}

		logAttrs[attrname.RetryNum] = attemptedRetries

		logEvt.Attrs(logAttrs).Msg(logMsg)
	}

	err = reqErr

	return
}

func delay(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-time.NewTimer(delay).C:
		return nil
	}
}

func New(opts ...OptionFn) (res http.RoundTripper, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of retryable HTTP client instrumenting of retrier; err: %w", err)
		}
	}()

	impl := &impl{
		retryBufLimit: ioext.DefaultApplicationBufferSize,
		maxRetries:    DefaultMaxRetries,
	}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.underlying == nil {
		err = errors.New("underlying round tripper not specified")

		return
	}

	if impl.retryBufLimit <= 0 && impl.maxRetries > 0 {
		err = errors.New("invalid value of retry buffer limit, can't be zero or negative")

		return
	}

	if impl.backoff == nil {
		err = errors.New("backoff strategy not specified whereas it expected")

		return
	}

	res = impl

	return
}
