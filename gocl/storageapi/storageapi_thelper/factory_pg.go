package storageapi_thelper

import (
	"context"
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
)

func NewPgFactory[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, opts ...FactoryOptionFn[TAPI]) factory.Factory[TAPI] {
	t.Helper()

	opts = append(opts, FactoryOption[TAPI]().StorageType(storage.PostgreSQLStorageType))

	return NewFactory(t, ctx, opts...)
}
