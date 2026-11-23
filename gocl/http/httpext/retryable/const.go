package retryable

import "rn.m.vo7.io/vo7io/m/gocl/http/httpext/retryable/retrier"

const (
	confExplanationPrefix = "retryable HTTP client"

	ConfFlenvSelfFlagPrefix = "http-client"

	DefaultMaxRetries     = retrier.DefaultMaxRetries
	DefaultMetricsEnabled = true
	DefaultTracingEnabled = false

	LogScope = "cl_http_retryable_client"
)
