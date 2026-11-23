package storageapi_thelper

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storage/storage_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

type FactoryOptionFn[TAPI storageapi.SQLAPI] func(*factbldr[TAPI])

func (o FactoryOptionFn[TAPI]) InstanceContextKey(val contextext.Key) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.instanceContextKey = val }
}

func (o FactoryOptionFn[TAPI]) StorageType(val storage.StorageType) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.storageType = val }
}

func (o FactoryOptionFn[TAPI]) APIConstructor(val factory.APIConstructorFn[TAPI]) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.apiConstructor = val }
}

func (o FactoryOptionFn[TAPI]) ConnConfEdit(val func(*testing.T, *storage.ConnConf)) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.connConfEdit = val }
}

func (o FactoryOptionFn[TAPI]) APIConfEdit(val func(*testing.T, *storageapi.Conf)) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.apiConfEdit = val }
}

func (o FactoryOptionFn[TAPI]) APIConf(val func(t *testing.T) storageapi.Conf) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.apiConf = val }
}

func (o FactoryOptionFn[TAPI]) ConnConf(val func(t *testing.T) storage.ConnConf) FactoryOptionFn[TAPI] {
	return func(f *factbldr[TAPI]) { o(f); f.connConf = val }
}

func FactoryOption[TAPI storageapi.SQLAPI]() FactoryOptionFn[TAPI] { return func(f *factbldr[TAPI]) {} }

type factbldr[TAPI storageapi.SQLAPI] struct {
	apiConf            func(t *testing.T) storageapi.Conf
	apiConfEdit        func(*testing.T, *storageapi.Conf)
	apiConstructor     factory.APIConstructorFn[TAPI]
	connConf           func(t *testing.T) storage.ConnConf
	connConfEdit       func(*testing.T, *storage.ConnConf)
	storageType        storage.StorageType
	instanceContextKey contextext.Key
}

func (s *factbldr[TAPI]) connConfEdited(t *testing.T) storage.ConnConf {
	t.Helper()

	connConf := s.connConf(t)
	if s.connConfEdit != nil {
		s.connConfEdit(t, &connConf)
	}

	return connConf
}

func newFactbldr[TAPI storageapi.SQLAPI](t *testing.T, opts ...FactoryOptionFn[TAPI]) *factbldr[TAPI] {
	t.Helper()

	bldr := &factbldr[TAPI]{instanceContextKey: factory.InstanceContextKey}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(bldr)
	}

	if bldr.connConf == nil {
		bldr.connConf = func(t *testing.T) storage.ConnConf {
			t.Helper()

			if bldr.storageType.IsUnspecified() {
				return storage_thelper.FlenvRootConnConf(t)
			}

			return storage_thelper.FlenvRootConnConfStorageTyped(t, bldr.storageType)
		}
	}

	if bldr.apiConf == nil {
		bldr.apiConf = func(t *testing.T) storageapi.Conf {
			t.Helper()

			return FlenvRootConf(t)
		}
	}

	if bldr.apiConstructor == nil {
		bldr.apiConstructor = func(ctx context.Context, b storageapi.SQLAPI) (TAPI, error) {
			var zero TAPI

			switch typed := b.(type) {
			case TAPI:
				return typed, nil
			default:
				return zero, fmt.Errorf("unable to cast storage API instance of type '%s' to required type '%s'", stringsext.TypeStringOfVal(b), stringsext.TypeStringFor[TAPI]())
			}
		}
	}

	return bldr
}

func NewFactory[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, opts ...FactoryOptionFn[TAPI]) factory.Factory[TAPI] {
	t.Helper()

	bldr := newFactbldr(t, opts...)

	connConf := bldr.connConfEdited(t)

	apiConf := bldr.apiConf(t)
	if bldr.apiConfEdit != nil {
		bldr.apiConfEdit(t, &apiConf)
	}

	res, err := factory.New(
		factory.Option[TAPI]().
			APIConf(apiConf).
			APIConstructor(bldr.apiConstructor).
			ConnConf(connConf).
			InstanceContextKey(bldr.instanceContextKey).
			Log(logext_thelper.FromContext(t, ctx)),
	)
	require.NoError(t, err, "unexpected error while creating instance of storage API factory")
	require.NotNil(t, res, "unexpected nil as a result of storage API factory creation")

	thelper.Cleanup(t, func() error { return res.Close(ctx) })

	return res
}
