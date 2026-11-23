package retryable

import (
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/retrier"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	ClientSSLCertPEM string `validate:"required_with=ClientSSLKeyPEM"`
	ClientSSLKeyPEM  string `validate:"required_with=ClientSSLCertPEM"`

	ClientName                  string
	TracingRequestURLPathFilter string

	Backoff backoff.ExponentialConfig

	SingleRoundTripTimeoutDur time.Duration `validate:"min=-1ns"`
	MaxRetries                int64         `validate:"gte=0"`
	RetryBufferLimit          int64         `validate:"gte=0"`

	DefaultRateLimitPerMinute int32 `validate:"gte=0"`

	MetricsEnabled               bool
	TracingEnabled               bool
	CancelImmediatelyOnRateLimit bool
}

func (c Conf) SingleRoundTripTimeout() timeout.Timeout {
	return timeout.Timeout(c.SingleRoundTripTimeoutDur)
}

func (c *Conf) BindFlagSetPrefix(prefix string, fls *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	if err := c.Backoff.BindFlagSetPrefix(prefix+"backoff", fls); err != nil {
		return err
	}

	fls.StringVar(
		&c.ClientSSLCertPEM,
		prefix+"ssl-client-cert",
		"",
		confExplanationPrefix+": SSL client certificate (PEM format)",
	)

	fls.StringVar(
		&c.ClientSSLKeyPEM,
		prefix+"ssl-client-key",
		"",
		confExplanationPrefix+": SSL client certificate key (PEM format)",
	)

	fls.Int64Var(
		&c.MaxRetries,
		prefix+"max-retries",
		retrier.DefaultMaxRetries,
		confExplanationPrefix+": limit of request retries",
	)

	fls.DurationVar(
		&c.SingleRoundTripTimeoutDur,
		prefix+"trip-timeout",
		httpext.DefaultClientRequestTimeout.Dur(),
		confExplanationPrefix+": single trip timeout (not includes backoff delay)",
	)

	fls.BoolVar(
		&c.CancelImmediatelyOnRateLimit,
		prefix+"cancel-immediately-on-rate-limit",
		false,
		confExplanationPrefix+": indicates whether to immediately cancel single trip (request) on exceeding rate limit"+
			"\n\ttrue - if rate limit reached, then single trip timeout not honored and trip cancellation occurs immediately"+
			"\n\tfalse - single trip timeout implies waiting on rate limiter too (if any and rate limit exceeding)",
	)

	fls.BoolVar(
		&c.MetricsEnabled,
		prefix+"metrics-enabled",
		DefaultMetricsEnabled,
		confExplanationPrefix+": indicates whether to emit request metrics",
	)

	fls.BoolVar(
		&c.TracingEnabled,
		prefix+"tracing-enabled",
		DefaultTracingEnabled,
		confExplanationPrefix+": indicates whether to emit request trace",
	)

	fls.StringVar(
		&c.ClientName,
		prefix+"name",
		"",
		confExplanationPrefix+": specifies name of this client; it is recommended to specify distinguished name",
	)

	fls.Int64Var(
		&c.RetryBufferLimit,
		prefix+"request-body-buffer-limit",
		ioext.DefaultApplicationBufferSize,
		confExplanationPrefix+": max. size of request body buffer used on retries",
	)

	fls.Int32Var(
		&c.DefaultRateLimitPerMinute,
		prefix+"default-rate-limit-per-minute",
		0,
		confExplanationPrefix+": indicates default requests rate limit expressed as count of requests per minute"+
			"\n\tspecified limit will be applied to all issued by this client requests"+
			"\n\t0 - means no limit",
	)

	fls.StringVar(
		&c.TracingRequestURLPathFilter,
		prefix+"tracing-request-url-path-filter",
		"",
		confExplanationPrefix+": specifies regular expression of URL path of request to be traced (not matched URL path means to skip request tracing)"+
			"\n\tignored if tracing not enabled"+
			"\n\tempty string - no filter",
	)

	return nil
}

func (c *Conf) Validate(v validation.Validator) error { return v.Struct(c) }

// [FlenvRootConfig] returns retryable HTTP client configuration as it would be the root configuration object.
//
// Default configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConfig() (Conf, error) {
	return FlenvMemberConfig("")
}

// [FlenvMemberConfig] returns retryable HTTP client configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConfig(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
