package factory

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type baseImpl[TFrom storageapi.SQLAPI] struct{ u Factory[TFrom] }

func (s *baseImpl[TFrom]) StorageType() storage.StorageType {
	return s.u.StorageType()
}

func (s *baseImpl[TFrom]) Close(ctx context.Context) error {
	return s.u.Close(ctx)
}

func (s *baseImpl[TFrom]) CloseLog(ctx context.Context, l logext.Logger) {
	s.u.CloseLog(ctx, l)
}

func (s *baseImpl[TFrom]) FromContext(ctx context.Context) (storageapi.SQLAPI, error) {
	return s.u.FromContext(ctx)
}

func (s *baseImpl[TFrom]) Base() FactorySQL { return s }

func (s *baseImpl[TFrom]) New(ctx context.Context) (storageapi.SQLAPI, func(), context.Context, error) {
	return s.u.New(ctx)
}

func (s *baseImpl[TFrom]) Context(parent context.Context, api storageapi.SQLAPI) context.Context {
	var (
		converted TFrom
		ok        bool
	)

	if converted, ok = api.(TFrom); !ok {
		panic(
			fmt.Errorf(
				"unable to convert value of type '%s' to type '%s'; %w",
				stringsext.TypeStringOfVal(api),
				stringsext.TypeStringFor[TFrom](),
				errs.ErrUnexpectedTypeOfValueInContext,
			),
		)
	}

	return s.u.Context(parent, converted)
}
