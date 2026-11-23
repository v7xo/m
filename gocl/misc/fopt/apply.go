package fopt

import "fmt"

// [Apply] invokes specified options opts on target t.
func Apply[T any, TOption ~func(*T) error](t *T, opts ...TOption) (err error) {
	for i := range opts {
		if opts[i] == nil {
			continue
		}

		if err = opts[i](t); err != nil {
			err = fmt.Errorf("error while applying options; %w", err)

			return
		}
	}

	return
}
