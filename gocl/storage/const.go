package storage

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

const (
	InfiniteTimeout = timeout.Infinite

	ConnConfFlenvSelfFlagPrefix     = "storage"
	MasterDBConfFlenvSelfFlagPrefix = "storage-master"

	DefaultConnConnectTimeout = time.Second * 5

	DefaultPostgreSQLMasterDBDatabase = "postgres"
	DefaultPostgreSQLMasterDBUser     = "postgres"
	DefaultPostgreSQLMasterDBPassword = "postgres"

	DefaultPoolConnIdleTime      = time.Minute * 10
	DefaultPoolConnMaxLifetime   = time.Hour
	DefaultPoolEnabled           = true
	DefaultPoolHealthCheckPeriod = time.Minute
	DefaultPoolMaxConns          = 64
	DefaultPoolMinConns          = 0

	DefaultPostgreSQLStmtCacheMode = StmtCacheModeDescribe
	DefaultStmtCacheMode           = StmtCacheModeStorageTypeDefault

	StorageTypeClickHouse = StorageType("ch")
	StorageTypeMySQL      = StorageType("mysql")
	PostgreSQLStorageType = StorageType("pg")
	TarantoolStorageType  = StorageType("tnt")

	DefaultPostgreSQLConnPort = 5432
	DefaultTarantoolConnPort  = 3301

	PoolHealthCheckPeriodConfigFlagName = "pool-healthcheck-period"

	FieldDbTagName = "db"
)
