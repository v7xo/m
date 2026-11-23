package retrier

import (
	"context"
	"net/http"

	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
)

type (
	NeedRetryFn        func(*http.Response) (bool, error)
	NeedRetryContextFn func(context.Context, *http.Response) (bool, error)
	RoundTripFn        = httpext.RoundTripFn
)
