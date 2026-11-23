package valonce

import (
	"context"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/value/mayberr"
)

type (
	Of[T any] interface {
		isValOnce(val T)

		Eval(ctx context.Context) (T, error)
		TryGet() (T, bool)
	}

	of[T any] struct {
		res atomic.Pointer[mayberr.Of[T]]
		fn  func(context.Context) (T, error)
		ch  chan struct{}
	}
)

func (v *of[T]) TryGet() (val T, ok bool) {
	if res := v.res.Load(); res != nil {
		switch {
		case res.HasErr():
			return
		}

		val, ok = res.MustVal(), true

		return
	}

	return
}

func (v *of[T]) Eval(ctx context.Context) (val T, err error) {
	if res := v.res.Load(); res != nil {
		val, err = res.Val()

		return
	}

	select {
	case _, ok := <-v.ch:
		if !ok {
			return v.res.Load().Val()
		}
	case <-ctx.Done():
		err = context.Cause(ctx)

		return
	}

	fnVal, fnErr := v.fn(ctx)
	if fnErr == nil {
		v.res.Store(mayberr.ValPointer(fnVal))
	} else {
		v.res.Store(mayberr.ErrPointer[T](fnErr))
	}

	close(v.ch)

	return v.res.Load().Val()
}

// nolint
func (*of[T]) isValOnce(T) {}

func New[T any](fn func(context.Context) (T, error)) Of[T] {
	res := &of[T]{fn: fn, ch: make(chan struct{}, 1)}

	res.ch <- struct{}{}

	return res
}

func Val[T any](v T) Of[T] {
	res := &of[T]{}
	res.res.Store(mayberr.ValPointer(v))

	return res
}

func Err[T any](err error) Of[T] {
	res := &of[T]{}
	res.res.Store(mayberr.ErrPointer[T](err))

	return res
}
