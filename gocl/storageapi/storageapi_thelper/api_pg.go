package storageapi_thelper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/envext"
	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
	hlp "rn.m.vo7.io/vo7io/m/gocl/storage/storage_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
)

func NewPgDB[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, factOpt ...FactoryOptionFn[TAPI]) (TAPI, factory.Factory[TAPI], storage.ConnConf) {
	t.Helper()

	originalConnConfEdited := newFactbldr(t, factOpt...).connConfEdited(t)

	masterConf := hlp.FlenvRootMasterDBConf(t)
	masterAPIFactOpts := append(
		factOpt,
		FactoryOption[TAPI]().
			ConnConf(func(t *testing.T) storage.ConnConf {
				t.Helper()

				cp := originalConnConfEdited

				cp.Database = masterConf.Database
				cp.User = masterConf.User
				cp.Password = masterConf.Password

				return cp
			}).
			ConnConfEdit(nil),
	)

	masterAPI, _ := NewPg(t, ctx, masterAPIFactOpts...)

	userConnConf := originalConnConfEdited
	userConnConf.Database = fmt.Sprintf("%s_test_%s_%s", userConnConf.Database, time.Now().Format("20060102150405"), uuidext.MustRand4BHex())

	var (
		err               error
		createStorageUser = envext.TestCreateStorageUser()
	)

	if createStorageUser {
		const roleTTL = time.Hour * 24 * 2

		// user name and password both same as generate DB name.
		userConnConf.User = userConnConf.Database
		userConnConf.Password = userConnConf.Database

		err = masterAPI.ExecSQLNonQuery(
			ctx,
			query.AnonNamed(
				fmt.Sprintf(
					`create role "%s" with superuser login password '%s' valid until '%s'`,
					userConnConf.User,
					userConnConf.Password,
					time.Now().Add(roleTTL).Format(time.DateOnly),
				),
			),
		)
		require.NoError(t, err, "unexpected error while creating new role/user '%s'", userConnConf.User)
	}

	err = masterAPI.ExecSQLNonQuery(ctx, query.AnonNamed(fmt.Sprintf(`create database "%s" with owner "%s"`, userConnConf.Database, userConnConf.User)))
	require.NoError(t, err, "unexpected error while creating new database '%s'", userConnConf.Database)

	if envext.DebugTest() {
		t.Cleanup(
			func() {
				t.Logf("WARNING: skip dropping new database '%s' (because of test debug mode - see env. variable '%s')", userConnConf.Database, envext.DebugTestEnvVar)
			},
		)
	} else {
		t.Cleanup(
			func() {
				dropPgTestUserDatabase(t, ctx, logext_thelper.FromContext(t, ctx), masterAPI, userConnConf.Database, userConnConf.User, !createStorageUser)
			},
		)
	}

	userAPI, userAPIFact := NewPg(
		t,
		ctx,
		append(
			factOpt,
			FactoryOption[TAPI]().
				ConnConf(func(_ *testing.T) storage.ConnConf { return userConnConf }).
				ConnConfEdit(nil),
		)...,
	)

	return userAPI, userAPIFact, userConnConf
}

func NewPg[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, factOpt ...FactoryOptionFn[TAPI]) (TAPI, factory.Factory[TAPI]) {
	t.Helper()

	fact := NewPgFactory(t, ctx, factOpt...)

	return NewStorageTyped[TAPI](t, ctx, fact, storage.PostgreSQLStorageType), fact
}

func dropPgTestUserDatabase(t *testing.T, ctx context.Context, log logging.Logger, api storageapi.SQLAPI, name, role string, skipRoleDrop bool) {
	t.Helper()

	type params struct {
		TargetDBName string `db:"target_db_name"`
	}

	msgPrologue := "unexpected error while dropping test database '" + name + "'; "

	terminateProcessesQuery, err := query.NewNamed(
		"test_helper_terminate_processes",
		`
select
	pg_terminate_backend(act.pid)
from
	pg_stat_activity act
where
	act.datname = @target_db_name::text
  	and pid <> pg_backend_pid()`,
		query.Option().AtSign(),
	)
	require.NoError(t, err, msgPrologue+"failed to create query to terminate backend(s)")

	err = api.ExecSQLNonQuery(ctx, terminateProcessesQuery, params{TargetDBName: name})
	require.NoError(t, err, msgPrologue+"failed to execute query")

	dropDatabaseQuery, err := query.NewNamed(
		"test_helper_drop_user_database",
		`drop database "`+name+`"`,
		query.Option().AtSign(),
	)
	require.NoError(t, err, msgPrologue+"failed to create query to drop test user database")

	err = api.ExecSQLNonQuery(ctx, dropDatabaseQuery)
	require.NoError(t, err, msgPrologue+"failed to execute query")

	log.Info(ctx).Msgf("test database '%s' successfully dropped", name)

	switch {
	case !skipRoleDrop:
		dropRoleQuery, err := query.NewNamed(
			"test_helper_drop_user_role",
			`drop role "`+role+`"`,
			query.Option().AtSign(),
		)
		require.NoError(t, err, msgPrologue+"failed to create query to drop test user role")

		err = api.ExecSQLNonQuery(ctx, dropRoleQuery)
		require.NoError(t, err, msgPrologue+"failed to execute query")

		log.Info(ctx).Msgf("test role '%s' successfully dropped", role)
	}
}
