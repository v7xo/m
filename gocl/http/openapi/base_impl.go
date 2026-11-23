package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type BaseImplementation struct {
	fopt.LogOpt

	// RequestContentLimit defines default request content length limit.
	// 0 means no limit.
	RequestContentLimit int
	// PreventErrDetailedMessage indicates to shorten error message emitted by [WriteErrorWithStatus] method (to prevent secure info disclosure).
	PreventErrDetailedMessage bool
}

// Echo request handler.
func (s *BaseImplementation) GetEcho(w http.ResponseWriter, r *http.Request, params EchoParams) {
	s.WriteJSON(
		w,
		r,
		EchoReply{
			BuildInfo: info.BuildInfoString(),
			ClientID:  info.UserAgentLong(),
			Text:      params.Text,
		},
	)
}

func (s *BaseImplementation) WriteJSON(w http.ResponseWriter, r *http.Request, val any) {
	s.WriteJSONWithStatus(w, r, val, http.StatusOK)
}

func (s *BaseImplementation) WriteJSONWithStatus(w http.ResponseWriter, r *http.Request, val any, status int) {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		s.WriteError(w, r, fmt.Errorf("error while marshaling payload as JSON; err: %w", err))

		return
	}

	w.Header().Set(contentTypeHeaderName, jsonContentTypeWithCharset)
	s.WriteBytes(w, r, jsonBytes, status)
}

func (s *BaseImplementation) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status := errs.TranslateStatus(err).HTTPStatus()
	s.WriteErrorWithStatus(w, r, err, status)
}

func (s *BaseImplementation) WriteErrorAndLog(w http.ResponseWriter, r *http.Request, err error) {
	status := errs.TranslateStatus(err).HTTPStatus()

	s.Log().
		Err(r.Context(), err).
		Int(attrname.StatusCode, status).
		Str(attrname.HTTPReqURL, r.RequestURI).
		Send()

	s.WriteErrorWithStatus(w, r, err, status)
}

func (s *BaseImplementation) WriteErrorWithStatus(w http.ResponseWriter, r *http.Request, err error, status int) {
	errModel := Error{
		Code:       status,
		CodeName:   http.StatusText(status),
		CustomCode: errs.TranslateCustomErrorCode(err),
	}

	if s.PreventErrDetailedMessage {
		errModel.Message = placeholder.ErrDetailedMessageOmitted
	} else {
		errModel.Message = err.Error()
	}

	w.Header().Set(contentTypeHeaderName, jsonContentTypeWithCharset)
	w.WriteHeader(status)

	jsonPayload, jsonMarshalErr := json.Marshal(errModel)
	if jsonMarshalErr != nil {
		contextext.MustLog(r.Context()).
			Err(r.Context(), jsonMarshalErr).
			Msgf("failed to marshal error payload as JSON; request URI: %s", r.RequestURI)

		return
	}

	if _, err := w.Write(jsonPayload); err != nil {
		contextext.MustLog(r.Context()).
			Err(r.Context(), err).
			Msgf(failedToWriteResponseMsgTemplate, r.RequestURI)
	}
}

func (s *BaseImplementation) WriteErrorWithStatusAndLog(w http.ResponseWriter, r *http.Request, err error, status int) {
	s.Log().
		Err(r.Context(), err).
		Int(attrname.StatusCode, status).
		Str(attrname.HTTPReqURL, r.RequestURI).
		Send()

	s.WriteErrorWithStatus(w, r, err, status)
}

func (s *BaseImplementation) WriteBytes(w http.ResponseWriter, r *http.Request, b []byte, statusCode int) {
	w.WriteHeader(statusCode)

	if len(b) == 0 {
		return
	}

	if _, err := w.Write(b); err != nil {
		contextext.MustLog(r.Context()).
			Err(r.Context(), err).
			Msgf(failedToWriteResponseMsgTemplate, r.RequestURI)
	}
}

func (s *BaseImplementation) ReadJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	if err := s.RequireRequestJSONContentType(r); err != nil {
		return err
	} else if _, err := s.RequireRequestContentLengthLimit(r); err != nil {
		return err
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.BadArgsWrapWith(err)
	}

	if len(bytes) == 0 {
		return errs.BadArgsWrapWith(errors.New("empty body"))
	}

	if err = json.Unmarshal(bytes, v); err != nil {
		return errs.BadArgsWrapWith(fmt.Errorf("failed to unmarshal from JSON; err: %w", err))
	}

	return nil
}

func (s *BaseImplementation) RequireRequestContentLengthLimit(r *http.Request) (int, error) {
	l, err := s.RequireRequestContentLength(r)
	if err != nil || s.RequestContentLimit == 0 {
		return 0, err
	}

	if l > s.RequestContentLimit {
		return 0, errs.BadArgsWrapWith(errors.New("request content length too large"))
	}

	return l, nil
}

func (*BaseImplementation) RequireRequestContentLength(r *http.Request) (int, error) {
	vals := r.Header.Values(contentLengthHeaderName)

	if len(vals) == 0 {
		return 0, errs.BadArgsWrapWith(errors.New("request doesn't contain content length header"))
	} else if len(vals) != 1 {
		return 0, errs.BadArgsWrapWith(errors.New("request content length header is ambiguous"))
	}

	res, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		return 0, errs.BadArgsWrapWith(fmt.Errorf("failed to parse content length header as integer; err: %w", err))
	} else if res < 0 {
		return 0, errs.BadArgsWrapWith(fmt.Errorf("request content length specified by its header is negative; err: %w", err))
	}

	return int(res), nil
}

func (*BaseImplementation) RequireRequestJSONContentType(r *http.Request) error {
	ok, err := httpext.HeaderDeclaresJSONContentType(r.Header)
	if err != nil {
		return err
	} else if !ok {
		return errs.BadArgsWrapWith(fmt.Errorf("request content type mismatches expected; expected: %s", jsonContentType))
	}

	return nil
}
