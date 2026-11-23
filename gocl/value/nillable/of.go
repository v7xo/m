package value

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Nillable[T byte | int | uint | int32 | uint32 | int64 | uint64 | string] struct {
	_      [0]*T
	val    T
	hasVal bool
}

func (v Nillable[T]) HasVal() bool { return v.hasVal }

func (v Nillable[T]) Val() T {
	if v.hasVal {
		return v.val
	}

	panic(fmt.Errorf("%w; struct is nillable (%T) and represents nil", errs.ErrNilValue, v))
}

func NillableVal[T byte | int | uint | int32 | uint32 | int64 | uint64 | string](val T) Nillable[T] {
	return Nillable[T]{val: val, hasVal: true}
}
