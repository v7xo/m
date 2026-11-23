package errs

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/status"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/errcode"
	clgrpcstatus "rn.m.vo7.io/vo7io/m/gocl/status/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

func TranslateGRPCStatus(err error) (s *status.Status, exactly bool) {
	return clgrpcstatus.Translate(err)
}

func TranslateStatus(err error) opstatus.Name { return clerr.TranslateStatus(err) }

func TranslateLevel(err error) level.Name {
	if err == nil {
		return level.Error
	}

	var asLevel level.AsLevel
	if errors.As(err, &asLevel) {
		return asLevel.Level()
	}

	return level.Error
}

func TranslateCustomErrorCode(err error) errcode.CustomErrorCode {
	if err == nil {
		return ""
	}

	var asCustomCode errcode.AsCustomErrorCode
	if errors.As(err, &asCustomCode) {
		return asCustomCode.AsCustomErrorCode()
	}

	switch st, ok := clgrpcstatus.Translate(err); ok {
	case true:
		if custErr := clgrpcstatus.CustomError(st); custErr != nil {
			return errcode.CustomErrorCode(custErr.GetCustomCode())
		}
	}

	return ""
}

// TranslateHTTPStatus returns error if status >= [http.StatusBadRequest].
func TranslateHTTPStatus(status int) error {
	if status < http.StatusBadRequest {
		return nil
	}

	switch status {
	case http.StatusBadRequest:
		return ErrBadArgs
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthenticated
	case http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusConflict:
		return ErrAlreadyExists
	case http.StatusInternalServerError:
		return ErrInternalServerError
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	}

	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		return fmt.Errorf("client error, non-success status code '%d (%s)'", status, http.StatusText(status))
	}

	return fmt.Errorf("non-success status code '%d (%s)'", status, http.StatusText(status))
}
