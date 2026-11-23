package progress

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn[T any] func(*of[T]) error

func (o OptionFn[T]) Capacity(val int) OptionFn[T] {
	return fopt.Chain(o, func(t *of[T]) { t.cap = val })
}

func (o OptionFn[T]) Handler(val func(T)) OptionFn[T] {
	return fopt.Chain(o, func(t *of[T]) { t.handler = val })
}

func (o OptionFn[T]) Stop(val context.Context) OptionFn[T] {
	return fopt.Chain(o, func(t *of[T]) { t.stop = val })
}

func Option[T any]() OptionFn[T] { return fopt.Nop[of[T]] }
