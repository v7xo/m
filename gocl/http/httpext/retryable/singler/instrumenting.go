package singler

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ratelimiter"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
	roundTripFn                  httpext.RoundTripFn
	roundTripper                 http.RoundTripper
	rateLimiter                  ratelimiter.Service
	sslCertPEM                   []byte
	sslKeyPEM                    []byte
	roundTripTimeout             timeout.Timeout
	cancelImmediatelyOnRateLimit bool
}

// RoundTrip implements [http.RoundTripper] interface.
func (s *impl) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx, ctxCancel := s.roundTripTimeoutContext(req.Context())

	defer ctxCancel()

	defer func() {
		if err == nil {
			return
		}

		logAttrs := make(attrname.Dict, 1) //nolint

		switch {
		case resp != nil:
			respHeadersBuf := bytes.NewBuffer(nil)
			if respHeadersWriteErr := resp.Header.Write(respHeadersBuf); respHeadersWriteErr != nil {
				logAttrs[attrname.ResponseHeaders] = "<failed to write response headers; err: " + respHeadersWriteErr.Error() + ">"
			} else {
				logAttrs[attrname.ResponseHeaders] = respHeadersBuf.String()
			}
		}
	}()

	if s.rateLimiter != nil && !contextext.NoRateLimit(ctx) {
		switch s.cancelImmediatelyOnRateLimit {
		case true:
			var rateLimitNotFiring bool

			if rateLimitNotFiring, _, err = s.rateLimiter.TryWaitNext(ctx); err != nil {
				return
			} else if !rateLimitNotFiring {
				err = errors.Join(fmt.Errorf("%w; rate limiter: %s", errs.ErrCallerRateLimit, s.rateLimiter.String()), context.Canceled)

				return
			}
		default:
			if _, err = s.rateLimiter.WaitNext(ctx); err != nil {
				return
			}
		}
	}

	if s.roundTripFn == nil {
		resp, err = s.roundTripper.RoundTrip(req)
	} else {
		resp, err = s.roundTripFn(req, s.roundTripper)
	}

	return
}

func (s *impl) roundTripTimeoutContext(parent context.Context) (context.Context, context.CancelFunc) {
	if s.roundTripTimeout.IsInfinite() {
		return parent, nopContextCancel
	}

	return context.WithTimeoutCause(
		parent,
		s.roundTripTimeout.Dur(),
		fmt.Errorf("single round-trip timeout reached; timeout: %s; err: %w", s.roundTripTimeout, context.DeadlineExceeded),
	)
}

func nopContextCancel() {}

func New(opts ...OptionFn) (res http.RoundTripper, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of retryable HTTP client single round-trip instrumenting; err: %w", err)
		}
	}()

	impl := &impl{roundTripTimeout: httpext.DefaultClientRequestTimeout}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	cli := &http.Client{}

	if len(impl.sslCertPEM) != 0 || len(impl.sslKeyPEM) != 0 {
		var cert tls.Certificate

		if cert, err = tls.X509KeyPair(impl.sslCertPEM, impl.sslKeyPEM); err != nil {
			err = fmt.Errorf("error while constructing X.509 key pair for specified client SSL certificate & key; err: %w", err)

			return
		}

		cli.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			},
		}
	}

	impl.roundTripper = httpext.ClientAsRoundTripper(cli)

	res = impl

	return
}
