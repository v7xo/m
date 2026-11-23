package contract

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/mayberr"
)

var defaultDescriber = &atomic.Pointer[mayberr.Of[Describer]]{}

func DefaultDescriber() (_ Describer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get default contract describer; %w", err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	holder := defaultDescriber.Load()
	if holder == nil {
		return nil, errors.New("describer not set yet")
	}

	return holder.Val()
}

func MustDefaultDescriber() Describer {
	res, err := DefaultDescriber()
	if err != nil {
		panic(err)
	}

	return res
}

func SetDefaultDescriber(d Describer, replace bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to set default describer; %w", err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	if d == nil {
		return fmt.Errorf("%w; discriber can't be nil", errs.ErrBadArgs)
	}

	newVal := mayberr.ValPointer(d)

	if replace {
		defaultDescriber.Swap(newVal)

		return
	}

	if !defaultDescriber.CompareAndSwap(nil, newVal) {
		return fmt.Errorf("%w; describer already set and replace not allowed by specified flag", errs.ErrAlreadyExists)
	}

	return
}

func DescribeFor[T any]() (Contract, error) { return Describe(reflect.TypeFor[T]()) }

func Describe(typ reflect.Type) (res Contract, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to describe type '%s'; %w", stringsext.TypeString(typ), err)
		}
	}()

	d, err := DefaultDescriber()
	if err != nil {
		return
	}

	ctx := contextext.AppGracefulEnd(nil)

	return d.Describe(ctx, typ)
}

func MustDescribeFor[T any]() Contract {
	res, err := DescribeFor[T]()
	if err != nil {
		panic(err)
	}

	return res
}

func MustDescribe(typ reflect.Type) Contract {
	res, err := Describe(typ)
	if err != nil {
		panic(err)
	}

	return res
}
