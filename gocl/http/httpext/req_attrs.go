package httpext

import (
	"context"
	"fmt"
	"net/http"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
)

type RequestAttrs attrname.Dict

func (v RequestAttrs) ToDict() attrname.Dict { return attrname.Dict(v) }

func (v RequestAttrs) URL() string {
	switch res := v[attrname.HTTPReqURL].(type) {
	case string:
		return res
	}

	return ""
}

func (v RequestAttrs) Method() string {
	switch res := v[attrname.HTTPReqMethod].(type) {
	case string:
		return res
	}

	return ""
}

func (v RequestAttrs) URLAuthority() string {
	switch res := v[attrname.HTTPReqURLAuthority].(type) {
	case string:
		return res
	}

	return ""
}

func (v RequestAttrs) URLPath() string {
	switch res := v[attrname.HTTPReqURLPath].(type) {
	case string:
		return res
	}

	return ""
}

func (v RequestAttrs) UserAgent() string {
	switch res := v[attrname.UserAgent].(type) {
	case string:
		return res
	}

	return ""
}

// RequestWithAttrs enriches specified request with context containing request logging attributes [RequestAttrs].
func RequestWithAttrs(req *http.Request, renew bool) (*http.Request, RequestAttrs, error) {
	ctx := req.Context()

	logAttrs, ok, err := contextext.Val[RequestAttrs](ctx, contextext.KeyHTTPRequestLogAttrs)
	if err != nil {
		return nil, nil, fmt.Errorf("error while enriching HTTP request object with log attributes; %w", err)
	}

	if ok && !renew {
		return req, logAttrs, nil
	}

	logAttrs = NewRequestAttrs(req)
	newReq := req.WithContext(context.WithValue(ctx, contextext.KeyHTTPRequestLogAttrs, logAttrs))

	return newReq, logAttrs, nil
}

func NewRequestAttrs(req *http.Request) RequestAttrs {
	res := make(RequestAttrs, 5) //nolint

	res[attrname.HTTPReqURL] = req.URL.Redacted()

	reqMethod, reqURLAuthority, reqURLPath := RequestMethodURLMetricLabels(req)

	res[attrname.HTTPReqMethod] = reqMethod
	res[attrname.HTTPReqURLAuthority] = reqURLAuthority
	res[attrname.HTTPReqURLPath] = reqURLPath
	res[attrname.UserAgent] = req.UserAgent()

	return res
}
