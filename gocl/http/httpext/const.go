package httpext

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

const (
	ServerConfigFlagPrefix = "http"

	DefaultListenHost = "0.0.0.0"
	DefaultListenPort = 8080

	DefaultClientRequestTimeout  = timeout.Timeout(time.Second * 7)
	DefaultServerShutdownTimeout = timeout.Timeout(time.Second * 17)

	RESTAPICallStyle = "rest"

	ContentTypeHeaderName   = "Content-Type"
	AuthorizationHeaderName = "Authorization"

	JSONContentType            = "application/json"
	JSONContentTypeWithCharset = "application/json; charset=utf-8"
	FormURLEncodedContentType  = "application/x-www-form-urlencoded"
)
