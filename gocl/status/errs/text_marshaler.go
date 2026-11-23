package errs

import (
	"encoding"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type TextError struct{ u error }

var (
	_ interface{ As(any) bool }     = &TextError{}
	_ interface{ Is(error) bool }   = &TextError{}
	_ interface{ Unwrap() []error } = &TextError{}
	_ error                         = &TextError{}
	_ encoding.TextMarshaler        = &TextError{}
)

func (e *TextError) MarshalText() (text []byte, err error) {
	if e == nil {
		return nil, nil
	}

	return []byte(e.u.Error()), nil
}

func (e *TextError) Error() string {
	if e == nil {
		return placeholder.Nil
	}

	return e.u.Error()
}

func (e *TextError) Unwrap() []error {
	if e == nil {
		return nil
	}

	c1, ok := e.u.(interface{ Unwrap() []error })
	if ok {
		return c1.Unwrap()
	}

	c2, ok := e.u.(interface{ Unwrap() error })
	if ok {
		return []error{c2.Unwrap()}
	}

	return nil
}

func (e *TextError) Is(t error) bool {
	if e == nil {
		return false
	}

	c, ok := e.u.(interface{ Is(error) bool })
	if ok {
		return c.Is(t)
	}

	return false
}

func (e *TextError) As(val any) bool {
	if e == nil {
		return false
	}

	c, ok := e.u.(interface{ As(any) bool })
	if ok {
		return c.As(val)
	}

	return false
}

func AsText(err error) *TextError {
	if ptrof.IsNil(err) {
		return nil
	}

	return &TextError{u: err}
}
