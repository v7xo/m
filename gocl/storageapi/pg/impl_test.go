//go:build test_postgres

package pg_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/storageapi_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_Ping(t *testing.T) {
	ctx := thelper.Pre(t)

	dbAPI, _ := storageapi_thelper.NewPg[storageapi.SQLAPI](t, ctx)
	err := dbAPI.Ping(ctx)
	require.NoError(t, err, "unexpected error while executing ping to DB")
}

func Test_CommandTimeout(t *testing.T) {
	hlp.PreLog(t)

	type (
		testCase struct {
			stdcase.CommonProps
		}
	)

	testCases := map[string]testCase{
		"001: ": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: context.DeadlineExceeded,
				},
			},
		},
	}

	for tcName, tc := range testCases {
		tcName, tc := tcName, tc

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				db_api, _ := storageapi_thelper.NewPg(
					t,
					ctx,
					storageapi_thelper.FactoryOption[storageapi.SQLAPI]().
						APIConfEdit(func(_ *testing.T, c *storageapi.Conf) { c.CommandTimeoutDur = time.Second * 2 }),
				)

				err := db_api.BeginTx(ctx)
				require.NoError(t, err)

				err = db_api.ExecSQLNonQuery(
					ctx,
					query.AnonNamed("select pg_sleep(:sleep_timeout);"),
					struct {
						SleepTimeout float32 `db:"sleep_timeout"`
					}{
						SleepTimeout: 4.0,
					},
				)
				tc.RequireConformError(t, err)

				require.Equal(t, false, db_api.IsOperable())

				err = db_api.ExecSQLNonQuery(ctx, query.AnonNamed("select 1"))
				require.ErrorIs(t, err, errs.ErrNotOperable)

				err = db_api.RollbackTx(ctx)
				require.ErrorIs(t, err, errs.ErrNotOperable)

				err = db_api.Ping(ctx)
				require.ErrorIs(t, err, errs.ErrNotOperable)
			},
		)
	}
}

func Test_ExecInTx(t *testing.T) {
	ctx := hlp.Pre(t)

	testTableName := fmt.Sprintf("test_table_%s", uuid_thelper.Rand4BHex(t))
	testValBeforeTx := int64(10)
	testValueInTx := 20
	dummyInTxErr := errors.New("this is dummy test error")

	initialQuery := query.AnonNamed(
		fmt.Sprintf(
			"create table %[1]s (val int8 not null); insert into %[1]s (val) select %[2]d",
			testTableName,
			testValBeforeTx,
		),
	)
	updateQuery := query.AnonNamed(
		fmt.Sprintf(
			"update %[1]s set val = %[2]d",
			testTableName,
			testValueInTx,
		),
	)
	selectQuery := query.AnonNamed(fmt.Sprintf("select val from %s", testTableName))

	dbAPI, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)
	err := dbAPI.ExecSQLNonQuery(ctx, initialQuery)
	require.NoError(t, err)

	err = dbAPI.ExecTxFn(
		ctx,
		func() error {
			inTxErr := dbAPI.ExecSQLNonQuery(ctx, updateQuery)
			if inTxErr != nil {
				return inTxErr
			}

			return dummyInTxErr
		},
	)
	require.ErrorIs(t, err, errs.ErrTxRollingback, "unexpected missing '%s' in error's chain", errs.ErrTxRollingback)
	require.ErrorIs(t, err, dummyInTxErr, "unexpected missing '%s' in error's chain", dummyInTxErr)

	var pgRes struct {
		Val pgtype.Int8 `db:"val"`
	}

	err = dbAPI.ExecSQLOneRow(ctx, selectQuery, &pgRes)
	require.NoError(t, err, "unexpected error while selecting one row from test table")

	require.True(t, pgRes.Val.Status == pgtype.Present, "unexpected mismatch of got one row")
	require.Equal(t, testValBeforeTx, pgRes.Val.Int, "unexpected mismatch of got one row")
}
