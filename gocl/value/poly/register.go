package poly

import (
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"
)

func Register[T any](opts ...contract.RegistrationOptionForIface[T]) error {
	return contract.Register(opts...)
}

func MustRegister[T any](opts ...contract.RegistrationOptionForIface[T]) struct{} {
	if err := Register(opts...); err != nil {
		panic(err)
	}

	return struct{}{}
}

func Options[T any]() contract.RegistrationOptionForFn[T] { return contract.RegistrationOptionFor[T]() }
