package fopt

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
)

// Pack creates a new option function which wraps specified options.
func Pack[T any, TOption ~func(*T) error](left TOption, right ...TOption) TOption {
	return func(t *T) error {
		if err := left(t); err != nil {
			return err
		}

		return Apply(t, right...)
	}
}

// PackSlice creates a new option function which wraps specified options.
func PackSlice[T any, TOption ~func(*T) error](left []TOption, right ...TOption) TOption {
	return func(t *T) error {
		if err := Apply(t, left...); err != nil {
			return err
		}

		return Apply(t, right...)
	}
}

// PrependFromContext creates a new option function which wraps specified opts with prepended context specific options. Consider, latter precede any
// specified by opts.
func PrependFromContext[T any, TOption ~func(*T) error](ctx context.Context, opts ...TOption) TOption {
	return func(t *T) error {
		if logSetter, ok := any(t).(LogSetter); ok {
			log, err := contextext.Log(ctx)
			if err != nil {
				return err
			}

			if err := logSetter.SetLog(log); err != nil {
				return err
			}
		}

		return Apply(t, opts...)
	}
}

// Chain creates a new option function that calls left option then right specified as function.
func Chain[T any, TOption ~func(*T) error](left TOption, right func(*T)) TOption {
	if left == nil && right == nil {
		return nil
	}

	return func(t *T) error {
		if left != nil {
			if err := left(t); err != nil {
				return err
			}
		}

		if right != nil {
			right(t)
		}

		return nil
	}
}

// ChainErr creates a new option function that calls left option then right specified as function.
func ChainErr[T any, TOption ~func(*T) error](left TOption, right func(*T) error) TOption {
	if left == nil && right == nil {
		return nil
	}

	return func(t *T) error {
		if left != nil {
			if err := left(t); err != nil {
				return err
			}
		}

		if right != nil {
			return right(t)
		}

		return nil
	}
}

func Concat[T any, TOption ~func(*T) error](first TOption, opts ...TOption) []TOption {
	if len(opts) == 0 {
		if first == nil {
			return nil
		}
	} else if first == nil {
		return opts
	}

	res := make([]TOption, len(opts)+1)

	copy(res[1:], opts)

	res[0] = first

	return res
}
