package retryable

import (
	"net/http"
)

type ObserveResultFn func(req *http.Request, resp *http.Response, err error)
