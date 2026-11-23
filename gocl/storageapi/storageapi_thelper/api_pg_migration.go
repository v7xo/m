package storageapi_thelper

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/storagemig_thelper"
)

func NewPgDBMigratedUp[TAPI storageapi.SQLAPI](t *testing.T, ctx context.Context, sources fs.FS, confFn func(*testing.T) storagemig.Conf, factOpt ...FactoryOptionFn[TAPI]) (TAPI, factory.Factory[TAPI], storage.ConnConf) {
	t.Helper()

	dbAPI, dbAPIFact, dbAPIConnConf := NewPgDB[TAPI](t, ctx, factOpt...)
	migrationSvc := storagemig_thelper.NewService(t, ctx, sources, confFn, dbAPI)

	err := migrationSvc.Initialize(ctx, false)
	require.NoError(t, err, "unexpected error while initializing new database '%s' for migrations to apply", dbAPIConnConf.Database)

	err = migrationSvc.Up(ctx, false)
	require.NoError(t, err, "unexpected error while applying migrations to new database '%s'", dbAPIConnConf.Database)

	status, err := migrationSvc.Status(ctx)
	require.NoError(t, err, "unexpected error while getting migrations status for new database '%s'", dbAPIConnConf.Database)

	dbAPI.Log().Info(ctx).Msgf("all migrations applied to database '%s'; migrations status: '%s'", dbAPIConnConf.Database, status)

	return dbAPI, dbAPIFact, dbAPIConnConf
}

func NewPgDBMigratedUpDown(t *testing.T, ctx context.Context, sources fs.FS, confFn func(*testing.T) storagemig.Conf, factOpt ...FactoryOptionFn[storageapi.SQLAPI]) {
	t.Helper()

	dbAPI, _, dbAPIConnCfg := NewPgDB[storageapi.SQLAPI](t, ctx, factOpt...)
	newDBName := dbAPIConnCfg.Database

	migrationSvc := storagemig_thelper.NewService(t, ctx, sources, confFn, dbAPI)

	err := migrationSvc.Initialize(ctx, false)
	require.NoError(t, err, "unexpected error while initializing new database '%s' for migrations to apply", newDBName)

	type stateDataRow struct {
		Name     string `db:"name"`
		NsName   string `db:"ns_name"`
		Kind     string `db:"kind"`
		OID      int64  `db:"oid"`
		OwnerOID int64  `db:"owner_oid"`
	}

	const stateDataQueryText = `
select
	cl.oid::int8 as "oid"
	,cl.relname::text as "name"
	,coalesce(ns.nspname::text, '<n/a>') as ns_name
	,cl.relkind::text as kind
	,cl.relowner::int8 as owner_oid
from
	pg_catalog.pg_class cl
left join
	pg_catalog.pg_namespace ns
on
	cl.relnamespace = ns.oid
where
	ns.nspname is null
	or ns.nspname::text not in ('information_schema', 'pg_catalog', 'pg_toast')
order by
	cl.oid asc`

	stateDataQuery, err := query.NewNamed(t.Name()+"_state_get", stateDataQueryText, query.Option().AtSign())
	require.NoError(t, err, "unexpected error while creating query for retrieving database state; database: %s", newDBName)

	stateDataBefore := []stateDataRow{}

	err = dbAPI.ExecSQL(ctx, stateDataQuery, &stateDataBefore)
	require.NoError(t, err, "unexpected error while retrieving database state before migrations apply; database: %s", newDBName)

	err = migrationSvc.Up(ctx, false)
	require.NoError(t, err, "unexpected error while applying migrations; database: %s", newDBName)

	status, err := migrationSvc.Status(ctx)
	require.NoError(t, err, "unexpected error while retrieving migrations status after migrations apply; database: %s", newDBName)

	dbAPI.Log().Info(ctx).Msgf("all migrations applied to database; database: %s; migrations status: %s", newDBName, status)

	err = migrationSvc.Down(ctx, false)
	require.NoError(t, err, "unexpected error while rolling back migrations; database: %s", newDBName)

	status, err = migrationSvc.Status(ctx)
	require.NoError(t, err, "unexpected error while retrieving migrations status after rolling back migrations; database: %s", newDBName)

	dbAPI.Log().Info(ctx).Msgf("all migrations rolled back; database: %s; migrations status: %s", newDBName, status)

	stateDataAfter := []stateDataRow{}

	err = dbAPI.ExecSQL(ctx, stateDataQuery, &stateDataAfter)
	require.NoError(t, err, "unexpected error while retrieving database state after migrations rolled back; database: %s", newDBName)

	require.ElementsMatch(t, stateDataBefore, stateDataAfter, "unexpected mismatch of database state after rolling back migrations (see details below); database: %s", newDBName)
}
