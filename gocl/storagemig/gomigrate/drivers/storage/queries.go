package storage

import (
	"time"

	db_q "rn.m.vo7.io/vo7io/m/gocl/storage/query"
)

type (
	lockSetQueryParams struct {
		LockIDHashCode int64 `db:"lock_id"`
	}

	metaInfoExistenceCheckQueryParams struct {
		TableSchema string   `db:"table_schema"`
		TableNames  []string `db:"table_names"`
	}

	versionQueryParams struct {
		CreatedAt         time.Time `db:"created_at"`
		ModifiedAt        time.Time `db:"modified_at"`
		DeployFingerprint int64     `db:"deploy_fingerprint"`
		Version           int32     `db:"version_num"`
	}

	deployQueryParams struct {
		CreatedAt      time.Time `db:"created_at"`
		ModifiedAt     time.Time `db:"modified_at"`
		AppRelease     string    `db:"app_release"`
		Fingerprint    int64     `db:"fingerprint"`
		IsDirty        bool      `db:"is_dirty"`
		IsUnregistered bool      `db:"is_unregistered"`
	}

	versionsNumsForDeployGetQueryParams struct {
		DeployFingerprint int64 `db:"deploy_fingerprint"`
	}

	versionsUnsetQueryParams struct {
		Versions          []int32 `db:"versions_nums"`
		DeployFingerprint int64   `db:"deploy_fingerprint"`
	}
)

const (
	metaInfoSchemeCreateQueryName     = `migration_meta_info_scheme_create`
	metaInfoSchemeCreateQueryTemplate = `CREATE SCHEMA "%s"`

	metaInfoVersionTableCreateQueryName     = `migration_meta_info_version_table_create`
	metaInfoVersionTableCreateQueryTemplate = `
CREATE TABLE "%s"."%s" (
	apply_order SERIAL PRIMARY KEY
	,deploy_fingerprint BIGINT NOT NULL
	,version_num INTEGER NOT NULL UNIQUE
	,created_at TIMESTAMPTZ NOT NULL
	,modified_at TIMESTAMPTZ NOT NULL
	,CHECK (deploy_fingerprint != 0)
	,CHECK (version_num > 0))
`

	metaInfoDeployTableCreateQueryName     = `migration_meta_info_deploy_table_create`
	metaInfoDeployTableCreateQueryTemplate = `
CREATE TABLE "%s"."%s" (
	apply_order SERIAL PRIMARY KEY
	,fingerprint BIGINT NOT NULL UNIQUE
	,is_dirty BOOLEAN NOT NULL
	,is_unregistered BOOLEAN NOT NULL
	,app_release VARCHAR(255) NOT NULL
	,created_at TIMESTAMPTZ NOT NULL
	,modified_at TIMESTAMPTZ NOT NULL
	,CHECK (fingerprint != 0))
`

	versionsNumsForDeployGetQueryName     = `migration_versions_nums_for_deploy_get`
	versionsNumsForDeployGetQueryTemplate = `
SELECT
	v.version_num
FROM
	"%[1]s"."%[2]s" v
	INNER JOIN
		"%[1]s"."%[3]s" d
	ON
		v.deploy_fingerprint = d.fingerprint
WHERE
	d.is_unregistered = false
	AND d.fingerprint = :deploy_fingerprint
ORDER BY
	v.apply_order ASC`

	versionInsertQueryName     = `migration_version_insert`
	versionInsertQueryTemplate = `
INSERT INTO "%s"."%s"
	(version_num
	,created_at
	,modified_at
	,deploy_fingerprint)
VALUES (
	:version_num
	,:created_at
	,:created_at
	,:deploy_fingerprint);
`

	versionExistsQueryName     = `migration_version_exists`
	versionExistsQueryTemplate = `
SELECT
	COALESCE(
		(SELECT true FROM "%s"."%s" WHERE version_num = :version_num LIMIT 1)
		,false)
`

	versionsUnsetQueryName     = `migration_versions_unset`
	versionsUnsetQueryTemplate = `
WITH
	del AS (
		DELETE FROM "%s"."%s"
		WHERE
			version_num = ANY(CAST(:versions_nums AS INTEGER[]))
			AND deploy_fingerprint = :deploy_fingerprint
		RETURNING version_num)
SELECT
	COUNT(*)
FROM
	del
`
	versionLatestGetQueryName     = `migration_version_latest_get`
	versionLatestGetQueryTemplate = `
SELECT
	v.apply_order
	,v.version_num
	,v.created_at
	,v.modified_at
	,v.deploy_fingerprint
	,d.app_release AS deploy_app_release
	,d.is_dirty AS is_deploy_dirty
FROM
	"%[1]s"."%[2]s" v
	INNER JOIN
		"%[1]s"."%[3]s" d
	ON
		v.deploy_fingerprint = d.fingerprint
WHERE
	d.is_unregistered = false
ORDER BY
	v.apply_order DESC
LIMIT 1
`

	deployRegisterInsertQueryName     = `migration_deploy_register_insert`
	deployRegisterInsertQueryTemplate = `
INSERT INTO "%s"."%s"
	(fingerprint
	,app_release
	,is_dirty
	,is_unregistered
	,created_at
	,modified_at)
VALUES (
	:fingerprint
	,:app_release
	,:is_dirty
	,false
	,:created_at
	,:created_at)
`
	deployRegisterUpdateQueryName     = `migration_deploy_register_update`
	deployRegisterUpdateQueryTemplate = `
UPDATE "%s"."%s" SET
	is_dirty = :is_dirty
	,is_unregistered = false
	,modified_at = :modified_at
	,app_release = :app_release
WHERE
	fingerprint = :fingerprint
`

	deployUnregisterQueryName     = `migration_deploy_unregister`
	deployUnregisterQueryTemplate = `
UPDATE "%s"."%s" SET
	is_unregistered = true
	,is_dirty = false
	,modified_at = :modified_at
WHERE
	fingerprint = :fingerprint
`

	registeredDeployExistsQueryName     = `migration_registered_deploy_exists`
	registeredDeployExistsQueryTemplate = `
SELECT
	COALESCE(
		(SELECT true FROM "%s"."%s" WHERE fingerprint = :fingerprint AND is_unregistered = false LIMIT 1)
		,false)
`

	deployGetQueryName     = `migration_deploy_get`
	deployGetQueryTemplate = `
SELECT
	apply_order
	,fingerprint
	,is_dirty
	,is_unregistered
	,created_at
	,modified_at
	,app_release
FROM
	"%s"."%s"
WHERE
	fingerprint = :fingerprint
`

	deployLatestRegisteredGetQueryName = `migration_deploy_latest_get`
	//nolint:unqueryvet
	deployLatestRegisteredGetQueryTemplate = `
SELECT
	d.*
	,COALESCE(v.version_num, 0) as version_num
	,COALESCE(v.apply_order, 0) as version_apply_order
	,(v.apply_order is null) as no_version
FROM (
	SELECT
		apply_order
		,fingerprint
		,is_dirty
		,is_unregistered
		,created_at
		,modified_at
		,app_release
	FROM
		"%[1]s"."%[2]s"
	WHERE
		is_unregistered = false
	ORDER BY
		apply_order DESC
	LIMIT 1) d
	LEFT JOIN
		"%[1]s"."%[3]s" v
	ON
		d.fingerprint = v.deploy_fingerprint
ORDER BY
	v.apply_order ASC
`

	migrationRunQueryName = `migration_run`
)

// nolint:gochecknoglobals
var (
	currentTxIDQuery            = db_q.MustNewNamed("migration_current_tx_id_get", `SELECT txid_current()`)
	lockSetQuery                = db_q.MustNewNamed("migration_advisory_lock_set", `SELECT pg_advisory_xact_lock(:lock_id)`)
	metaInfoExistenceCheckQuery = db_q.MustNewNamed(
		"migration_meta_info_existence_check",
		`
SELECT
	(SELECT
		COUNT(*)
	FROM
		"information_schema"."tables"
	WHERE
		table_schema = :table_schema AND table_name = ANY(cast(:table_names as text[]))) as "table_count"
	,COALESCE(
		(SELECT
			true
		FROM
			"information_schema"."schemata"
		WHERE "schema_name" = :table_schema
		LIMIT 1)
		,false) as "schema_exists"
`,
	)
)
