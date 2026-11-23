package tracing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
	underlying                  http.RoundTripper
	requestURLPathFilter        *regexp.Regexp
	requestURLPathFilterPattern string
}

// RoundTrip implements [http.RoundTripper] interface.
func (s *impl) RoundTrip(req *http.Request) (*http.Response, error) {
	req, logAttrs, logAttrsErr := httpext.RequestWithAttrs(req, false)
	if logAttrsErr != nil {
		s.Log().
			Err(req.Context(), logAttrsErr).
			Msg("HTTP tracing; error(s) occurred")

		return nil, logAttrsErr
	}

	if !s.traceRequest(req) {
		return s.underlying.RoundTrip(req)
	}

	reqHeaderCp := req.Header.Clone()
	headerBuf := bytes.NewBuffer(nil)

	if _, ok := reqHeaderCp[httpext.AuthorizationHeaderName]; ok {
		reqHeaderCp.Set(httpext.AuthorizationHeaderName, placeholder.PIIRedacted)
	}

	if reqHeaderErr := reqHeaderCp.Write(headerBuf); reqHeaderErr != nil {
		logAttrs[attrname.RequestHeaders] = "<failed to write request headers; " + reqHeaderErr.Error() + ">"
	} else {
		logAttrs[attrname.RequestHeaders] = headerBuf.String()
	}

	allErrs := []error{}

	if req.Body != nil {
		bodyBuf := bytes.NewBuffer(nil)

		switch _, reqBodyErr := io.Copy(bodyBuf, req.Body); reqBodyErr != nil {
		case true:
			allErrs = append(allErrs, fmt.Errorf("failed to copy request body to buffer for subsequent request tracing; %w", reqBodyErr))
			logAttrs[attrname.RequestBody] = "<failed to copy request body; " + reqBodyErr.Error() + ">"
		default:
			if reqBodyCloseErr := req.Body.Close(); reqBodyCloseErr != nil {
				allErrs = append(allErrs, fmt.Errorf("failed to close request body after reading it to buffer for subsequent request tracing; %w", reqBodyCloseErr))
			} else {
				req.Body = io.NopCloser(bodyBuf)
			}

			logAttrs[attrname.RequestBody] = bodyBuf.String()
		}
	} else {
		logAttrs[attrname.RequestBody] = placeholder.Nil
	}

	ctx := req.Context()

	if len(allErrs) != 0 {
		s.Log().
			Err(ctx, errors.Join(allErrs...)).
			Attrs(logAttrs.ToDict()).
			Msg("HTTP tracing; error(s) occurred")

		return nil, fmt.Errorf("error(s) occurred during collecting HTTP trace data (before fulfilling request); %w", errors.Join(allErrs...))
	}

	resp, respErr := s.underlying.RoundTrip(req)
	if respErr != nil {
		allErrs = append(allErrs, fmt.Errorf("round trip response indicates error; %w", respErr))
	}

	switch resp {
	case nil:
		logAttrs[attrname.ResponseHeaders] = headerBuf.String()
		logAttrs[attrname.ResponseBody] = placeholder.Nil
	default:
		headerBuf = bytes.NewBuffer(nil)

		if respHeaderErr := resp.Header.Write(headerBuf); respHeaderErr != nil {
			logAttrs[attrname.ResponseHeaders] = "<failed to write response headers; " + respHeaderErr.Error() + ">"
		} else {
			logAttrs[attrname.ResponseHeaders] = headerBuf.String()
		}

		if resp.Body != nil {
			bodyBuf := bytes.NewBuffer(nil)

			switch _, respBodyErr := io.Copy(bodyBuf, resp.Body); respBodyErr != nil {
			case true:
				allErrs = append(allErrs, fmt.Errorf("failed to copy response body to buffer for subsequent tracing; %w", respBodyErr))
				logAttrs[attrname.ResponseBody] = "<failed to copy response body; " + respBodyErr.Error() + ">"
			default:
				if respBodyCloseErr := resp.Body.Close(); respBodyCloseErr != nil {
					allErrs = append(allErrs, fmt.Errorf("failed to close response body after reading it to buffer for subsequent tracing; %w", respBodyCloseErr))
				}

				resp.Body = io.NopCloser(bodyBuf)
				logAttrs[attrname.ResponseBody] = bodyBuf.String()
			}
		} else {
			logAttrs[attrname.ResponseBody] = placeholder.Nil
		}
	}

	if len(allErrs) != 0 {
		s.Log().
			Err(ctx, errs.Join(allErrs...)).
			Attrs(logAttrs.ToDict()).
			Msg("HTTP tracing; error(s) occurred")

		return resp, errs.Join(errors.New("error(s) occurred during collecting HTTP trace data (see logs for details)"), respErr)
	}

	s.Log().
		Trace(ctx).
		Attrs(logAttrs.ToDict()).
		Msg("HTTP tracing; no error(s) occurred")

	return resp, nil
}

func (s *impl) traceRequest(req *http.Request) bool {
	if s.requestURLPathFilter == nil || req.URL == nil || s.requestURLPathFilter.MatchString(req.URL.Path) {
		return true
	}

	return false
}

func New(opts ...OptionFn) (res http.RoundTripper, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of retryable HTTP client tracing instrumenting; %w", err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.underlying == nil {
		err = errors.New("underlying round tripper not specified")

		return
	}

	switch {
	case impl.requestURLPathFilterPattern != "":
		if impl.requestURLPathFilter, err = regexp.Compile(impl.requestURLPathFilterPattern); err != nil {
			err = fmt.Errorf("failed to parse request URL path filter; expression: %s; %w", impl.requestURLPathFilterPattern, err)

			return
		}
	}

	res = impl

	return
}
