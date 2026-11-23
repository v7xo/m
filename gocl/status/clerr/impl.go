package clerr

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/status"

	clpbt "rn.m.vo7.io/vo7io/m/gocl/pbt/types"
	"rn.m.vo7.io/vo7io/m/gocl/status/errcode"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

type impl struct {
	orig *impl

	inner error

	statusName opstatus.Name
	level      level.Name

	baseMsg       string
	finalMsg      string
	customErrCode errcode.CustomErrorCode
}

var _ Error = &impl{}

func (e *impl) AsCustomErrorCode() errcode.CustomErrorCode { return e.customErrCode }
func (e *impl) String() string                             { return e.finalMsg }
func (e *impl) Error() string                              { return e.finalMsg }
func (e *impl) Unwrap() error                              { return e.inner }
func (e *impl) WithLevel(l level.Name) Error               { return e.WithStatusLevel(e.statusName, l) }
func (e *impl) WithStatus(st opstatus.Name) Error          { return e.WithStatusLevel(st, e.level) }

// Status implements [opstatus.StatusName].
func (e *impl) StatusName() opstatus.Name { return e.statusName }

// Level implements [level.AsLevel].
func (e *impl) Level() level.Name { return e.level }

//nolint:errorlint
func (e *impl) Is(t error) bool {
	if tStricted, ok := t.(*impl); ok && tStricted != nil {
		return e.orig == tStricted.orig
	}

	return false
}

func (e *impl) WithStatusLevel(s opstatus.Name, l level.Name) Error {
	if e == nil || (e.statusName == s && e.level == l) {
		return e
	}

	res := &impl{
		baseMsg:    e.baseMsg,
		inner:      e.inner,
		level:      l,
		orig:       e.orig,
		statusName: s,
	}

	res.setFinalMsg()

	return res
}

func (e *impl) GRPCStatus() *status.Status {
	if e == nil {
		return nil
	}

	stat := e.statusName.GRPCStatus()

	if e.AsCustomErrorCode() != "" {
		statWithDetails, err := stat.
			WithDetails(
				&clpbt.Error{
					CustomCode: string(e.customErrCode),
				},
			)
		if err == nil {
			stat = statWithDetails
		}
	}

	return stat
}

func (e *impl) setFinalMsg() {
	if e == nil {
		return
	}

	if (e.statusName.IsFail() || e.statusName.IsUnspecified()) && e.level == level.Error {
		e.finalMsg = e.baseMsg

		return
	}

	switch {
	case len(e.baseMsg) == 0:
		if e.statusName.IsUnspecified() {
			e.finalMsg = fmt.Sprintf("[ %s ]", e.level)
		} else {
			e.finalMsg = fmt.Sprintf("[ %s %s ]", e.level, e.statusName)
		}
	default:
		if e.statusName.IsUnspecified() {
			e.finalMsg = fmt.Sprintf("[ %s ] %s", e.level, e.baseMsg)
		} else {
			e.finalMsg = fmt.Sprintf("[ %s %s ] %s", e.level, e.statusName, e.baseMsg)
		}
	}
}

func Errf(msg string, a ...any) Error {
	return ErrStatusLevelf(opstatus.Unspecified, level.Error, msg, a...)
}

func ErrStatusf(s opstatus.Name, msg string, a ...any) Error {
	return ErrStatusLevelf(s, level.Error, msg, a...)
}

func ErrStatusCustomCodef(s opstatus.Name, msg string, c errcode.CustomErrorCode, a ...any) Error {
	return ErrLevelCustomCodef(s, level.Error, msg, c, a...)
}

func ErrLevelCustomCodef(s opstatus.Name, l level.Name, msg string, c errcode.CustomErrorCode, a ...any) Error {
	var (
		baseMsg        string
		stdErrf, inner error
	)

	if len(msg) != 0 {
		stdErrf = fmt.Errorf(msg, a...)
		baseMsg = stdErrf.Error()
		inner = errors.Unwrap(stdErrf)

		if inner != nil && s.IsUnspecified() {
			s = TranslateStatus(inner)
		}
	}

	e := &impl{
		baseMsg:       baseMsg,
		inner:         inner,
		level:         l,
		statusName:    s,
		customErrCode: c,
	}

	e.orig = e

	e.setFinalMsg()

	return e
}

func ErrStatusLevelf(s opstatus.Name, l level.Name, msg string, a ...any) Error {
	var (
		baseMsg        string
		stdErrf, inner error
	)

	if len(msg) != 0 {
		stdErrf = fmt.Errorf(msg, a...)
		baseMsg = stdErrf.Error()
		inner = errors.Unwrap(stdErrf)

		if inner != nil && s.IsUnspecified() {
			s = TranslateStatus(inner)
		}
	}

	e := &impl{
		baseMsg:    baseMsg,
		inner:      inner,
		level:      l,
		statusName: s,
	}

	e.orig = e

	e.setFinalMsg()

	return e
}

func Wrap(inner error) Error {
	return WrapStatusLevel(opstatus.Unspecified, level.Error, inner)
}

func WrapStatus(s opstatus.Name, inner error) Error {
	return WrapStatusLevel(s, level.Error, inner)
}

func WrapLevel(l level.Name, inner error) Error {
	return WrapStatusLevel(opstatus.Unspecified, l, inner)
}

//nolint:errorlint
func WrapStatusLevel(s opstatus.Name, l level.Name, inner error) Error {
	if inner == nil {
		return nil
	}

	if innerStricted, ok := inner.(Error); ok && innerStricted.Level() == l && innerStricted.StatusName() == s {
		return innerStricted
	}

	baseMsg := inner.Error()

	if s.IsUnspecified() {
		s = TranslateStatus(inner)
	}

	e := &impl{
		baseMsg:    baseMsg,
		inner:      inner,
		level:      l,
		statusName: s,
	}

	e.orig = e

	e.setFinalMsg()

	return e
}

func Wrapf(inner error, msg string, a ...any) Error {
	return WrapStatusLevelf(opstatus.Unspecified, level.Error, inner, msg, a...)
}

func WrapStatusf(s opstatus.Name, inner error, msg string, a ...any) Error {
	return WrapStatusLevelf(s, level.Error, inner, msg, a...)
}

func WrapLevelf(l level.Name, inner error, msg string, a ...any) Error {
	return WrapStatusLevelf(opstatus.Unspecified, l, inner, msg, a...)
}

//nolint:errorlint
func WrapStatusLevelf(s opstatus.Name, l level.Name, inner error, msg string, a ...any) Error {
	if len(msg) == 0 {
		return WrapStatusLevel(s, l, inner)
	}

	if inner == nil {
		return ErrStatusLevelf(s, l, msg, a...)
	}

	var (
		baseMsg        = fmt.Sprintf(msg, a...)
		innerErrString = inner.Error()
	)

	if innerErrString != "" {
		if len(baseMsg) > 0 {
			baseMsg += "; "
		}

		baseMsg += innerErrString
	}

	if s.IsUnspecified() {
		s = TranslateStatus(inner)
	}

	e := &impl{
		baseMsg:    baseMsg,
		inner:      inner,
		level:      l,
		statusName: s,
	}

	e.orig = e

	e.setFinalMsg()

	return e
}
