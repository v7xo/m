package factory

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

var _ = errs.ErrStorageAPIMissInContext

type FactorySQL = Factory[storageapi.SQLAPI]

type Factory[TAPI storageapi.SQLAPI] interface {
	StorageType() storage.StorageType

	Close(ctx context.Context) error

	CloseLog(ctx context.Context, l logext.Logger)

	// New returns instance of API accordingly with following:
	//
	// If ctx contains instance of API, that instance will be returned.
	// outCtx will be the same as ctx, closer will be non-nil NOP-function.
	//
	// If ctx doesn't contain instance of API, new instance will be created and associated with outCtx.
	New(ctx context.Context) (api TAPI, closer func(), outCtx context.Context, err error)

	// Context creates a new context associated with specified instance of API.
	Context(parent context.Context, api TAPI) context.Context

	// FromContext requires instance of API from the specified context.
	//
	// If no API provided by specified context, a wrapped error [errs.ErrStorageAPIMissInContext] returned.
	FromContext(ctx context.Context) (TAPI, error)

	// Base returns this factory as usable with components requiring factory that creates API of type [storageapi.SQLAPI].
	Base() FactorySQL
}
