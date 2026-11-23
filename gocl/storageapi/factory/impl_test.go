//go:build test_postgres

package factory_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/storageapi_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

type (
	dbAPI1Iface interface{ storageapi.SQLAPI }
	dbAPI1      struct{ storageapi.SQLAPI }
)

var (
	_ dbAPI1Iface = dbAPI1{}
	_ dbAPI1Iface = &dbAPI1{}
)

func (v *dbAPI1) SetBaseAPI(a storageapi.SQLAPI) error {
	v.SQLAPI = a

	return nil
}

func Test_Custom(t *testing.T) {
	ctx, log := thelper.PreLog(t)

	conf := storageapi_thelper.FlenvRootClientConf(t)

	factory, err := factory.New(
		factory.Option[dbAPI1]().
			APIConf(conf.API).
			APIConstructor(factory.CustomAPIConstructor).
			ConnConf(conf.Conn).
			Log(log),
	)
	require.NoError(t, err)

	thelper.Cleanup(t, func() error { return factory.Close(ctx) })

	dbAPI1, _, ctx, err := factory.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, dbAPI1)
	require.NotNil(t, dbAPI1.SQLAPI)

	thelper.Cleanup(t, func() error { return dbAPI1.Close(ctx) })

	err = dbAPI1.Ping(ctx)
	require.NoError(t, err, "error while making PING command on storage API")

	baseFactory := factory.Base()
	require.NotNil(t, baseFactory)

	baseAPI1, _, ctx, err := baseFactory.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, baseAPI1)

	thelper.Cleanup(t, func() error { return baseAPI1.Close(ctx) })

	baseAPI2, _, ctx, err := baseFactory.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, baseAPI2)

	thelper.Cleanup(t, func() error { return baseAPI2.Close(ctx) })

	require.Equal(t, baseAPI2, baseAPI1)

	ctx = baseFactory.Context(ctx, baseAPI1)

	baseAPI1FromContext, err := baseFactory.FromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, baseAPI1FromContext)
	require.Equal(t, baseAPI1, baseAPI1FromContext)
	require.Equal(t, baseAPI2, baseAPI1FromContext)

	dbAPI1FromContext, err := factory.FromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, dbAPI1FromContext)
	require.Equal(t, baseAPI1FromContext, dbAPI1FromContext) // same context
	require.Equal(t, dbAPI1, dbAPI1FromContext)
}
