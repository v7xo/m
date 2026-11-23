package storagemig

import (
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

const (
	ConfFlenvSelfFlagPrefix = "storage-migration"

	GoMigrateLib = types.GoMigrateLib
	DefaultLib   = types.GoMigrateLib

	PostgreSQLDBDriverName     = types.StorageDriverName("postgres")
	FileSystemSourceDriverName = types.SourceDriverName("file-system")

	DefaultStepRunStorageTimeout     = timeout.Infinite
	DefaultMetaInfoDeployObjectName  = "__migration_deploy"
	DefaultMetaInfoSchemaName        = "public"
	DefaultMetaInfoVersionObjectName = "__migration_version"

	UnassignedDeployFingerprint  = types.UnassignedDeployFingerprint
	UnassignedSourcesFingerprint = types.UnassignedSourcesFingerprint
)
