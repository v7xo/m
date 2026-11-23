package storageapi_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	fact "rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func NewStorageTyped[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, f fact.Factory[TAPI], st storage.StorageType) TAPI {
	t.Helper()

	api, _, _, err := f.New(ctx) // nolint:staticcheck
	require.NoError(t, err)
	require.NotNil(t, api)

	thelper.Cleanup(t, func() error { return api.Close(ctx) })

	require.Equal(t, st, api.StorageType(), "unexpected storage type")

	return api
}
