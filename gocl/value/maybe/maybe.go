package maybe

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Of[T any] struct {
	_       [0]*T
	val     T
	notMiss bool
}

func (i Of[T]) NotMiss() bool { return i.notMiss }

func (i Of[T]) Miss() bool { return !i.notMiss }

func (v Of[T]) Val() T {
	if v.notMiss {
		return v.val
	}

	panic(errs.ErrMissingValue)
}

func (v Of[T]) ValOrDefault() (res T) {
	if v.notMiss {
		return v.val
	}

	return
}

func (v Of[T]) AsAny() Of[any] {
	if v.notMiss {
		return Of[any]{val: v.val, notMiss: true}
	}

	return Of[any]{}
}

func (v Of[T]) String() string {
	if v.notMiss {
		return fmt.Sprintf("%v", v.val)
	}

	return placeholder.Missing
}

func Val[T any](val T) Of[T] {
	return Of[T]{val: val, notMiss: true}
}

func ValErr[T any](val T, err error) (Of[T], error) {
	if err == nil {
		return Of[T]{val: val, notMiss: true}, nil
	}

	return Of[T]{}, err
}
