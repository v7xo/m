//go:build test_postgres

package pool_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	hlp_api "rn.m.vo7.io/vo7io/m/gocl/storageapi/storageapi_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Ping(t *testing.T) {
	ctx := thelper.Pre(t)

	db_api, _ := hlp_api.NewPg[storageapi.SQLAPI](t, ctx)
	err := db_api.Ping(ctx)
	require.NoError(t, err, "unexpected error while executing ping to storage")
}
