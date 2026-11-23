package storagemig_thelper

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
)

func Up(t *testing.T, ctx context.Context, sources fs.FS, confFn func(*testing.T) storagemig.Conf, api storageapi.SQLAPI) storagemig.Service {
	t.Helper()

	svc := NewService(t, ctx, sources, confFn, api)

	err := svc.Up(ctx, false)
	require.NoError(t, err, "unexpected error while making up of storage (SQL) migrations")

	return svc
}
