package storagemig_thelper

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/factory"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func NewService(t *testing.T, ctx context.Context, sources fs.FS, confFn func(*testing.T) storagemig.Conf, api storageapi.SQLAPI) storagemig.Service {
	t.Helper()

	var conf storagemig.Conf

	if confFn == nil {
		conf = FlenvRootConf(t)
	} else {
		conf = confFn(t)
	}

	svc, svcClose, err := factory.NewServiceSQLAPI(ctx, sources, conf, api)
	require.NoError(t, err, UnexpecteErrorOnServiceCreationMsg)
	require.NotNil(t, svc, ServiceNotCreatedMsg)
	require.NotNil(t, svcClose, ServiceCloseFnNilMsg)

	thelper.Cleanup(t, func() error { return svcClose(ctx) })

	return svc
}

func NewServiceSQLAPIConf(t *testing.T, ctx context.Context, sources fs.FS, cfg storagemig.Conf, apiCfg storageapi.Conf, connCfg storage.ConnConf, log logext.Logger) storagemig.Service {
	t.Helper()

	svc, svcClose, err := factory.NewServiceSQLAPIConf(ctx, sources, cfg, apiCfg, connCfg, log)
	require.NoError(t, err, UnexpecteErrorOnServiceCreationMsg)
	require.NotNil(t, svc, ServiceNotCreatedMsg)
	require.NotNil(t, svcClose, ServiceCloseFnNilMsg)

	t.Cleanup(
		func() {
			if closeErr := svcClose(ctx); closeErr != nil {
				t.Log("[ERROR] error while calling close func of DB migration service")

				if !t.Failed() {
					t.Fail()
				}
			}
		},
	)

	return svc
}
