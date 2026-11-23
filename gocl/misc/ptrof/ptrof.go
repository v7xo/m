package ptrof

import (
	"reflect"
)

//go:fix inline
func Of[T any](v T) *T {
	return new(v)
}

func OfErr[T any](v T, err error) (*T, error) {
	return &v, err
}

func Eq[T comparable](a, b *T) bool {
	if a == nil {
		return b == nil
	} else if b == nil {
		return false
	}

	return *a == *b
}

func Copy[T any, TPointer interface{ *T }](val TPointer) TPointer {
	if val == nil {
		return nil
	}

	cp := *val

	return &cp
}

func ValOrDef[T any](a *T) (res T) {
	if a == nil {
		return
	}

	res = *a

	return
}

func IsNil(val any) bool {
	if val == nil {
		return true
	}

	reflVal := reflect.ValueOf(val)

	// nolint:exhaustive
	switch reflVal.Kind() {
	case
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		if reflVal.IsNil() {
			return true
		}
	}

	return false
}
