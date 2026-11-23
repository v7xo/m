package contract

import (
	"errors"
	"fmt"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/mayberr"
)

var defaultRegistrar = &atomic.Pointer[mayberr.Of[Registrar]]{}

func DefaultRegistrar() (_ Registrar, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get default registrar; %w", err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	holder := defaultRegistrar.Load()
	if holder == nil {
		return nil, errors.New("registrar not set yet")
	}

	return holder.Val()
}

func MustDefaultRegistrar() Registrar {
	res, err := DefaultRegistrar()
	if err != nil {
		panic(err)
	}

	return res
}

func SetDefaultRegistrar(r Registrar, replace bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to set default registrar; %w", err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	if r == nil {
		return fmt.Errorf("%w; registrar can't be nil", errs.ErrBadArgs)
	}

	newVal := mayberr.ValPointer(r)

	if replace {
		defaultRegistrar.Swap(newVal)

		return
	}

	if !defaultRegistrar.CompareAndSwap(nil, newVal) {
		return fmt.Errorf("%w; registrar already set and replace not allowed by specified flag", errs.ErrAlreadyExists)
	}

	return
}

func init() {
	d, err := NewDescriber()
	if err != nil {
		defaultDescriber.Swap(mayberr.ErrPointer[Describer](err))
		defaultRegistrar.Swap(mayberr.ErrPointer[Registrar](err))

		return
	}

	defaultDescriber.Swap(mayberr.ValPointer(d))

	r, err := NewRegistrar(RegistrarOption().WithDescriber(d))
	if err != nil {
		defaultRegistrar.Swap(mayberr.ErrPointer[Registrar](err))

		return
	}

	defaultRegistrar.Swap(mayberr.ValPointer(r))
}
