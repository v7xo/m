package mayberr

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Of[T any] struct {
	_           [0]*T
	val         T
	err         error
	initialized bool
}

func (m Of[TValue]) HasErr() bool {
	if m.initialized {
		return m.err != nil
	}

	panic(errs.ErrUninitializedValue)
}

func (m Of[TValue]) Err() error {
	if m.initialized {
		return m.err
	}

	panic(errs.ErrUninitializedValue)
}

func (m Of[TValue]) Val() (v TValue, err error) {
	if m.initialized {
		v, err = m.val, m.err

		return
	}

	panic(errs.ErrUninitializedValue)
}

func (m Of[TValue]) MustVal() (v TValue) {
	v, err := m.Val()
	if err == nil {
		return
	}

	panic(err)
}

func (m Of[TValue]) Valid() bool {
	return m.initialized
}

func Val[T any](v T) Of[T] {
	return Of[T]{val: v, initialized: true}
}

func Err[T any](err error) Of[T] {
	return Of[T]{err: err, initialized: true}
}

func ValPointer[T any](v T) *Of[T] {
	return &Of[T]{val: v, initialized: true}
}

func ErrPointer[T any](err error) *Of[T] {
	return &Of[T]{err: err, initialized: true}
}
