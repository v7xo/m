package storage_thelper

import (
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

func FlenvRootPgConnConf(t *testing.T) storage.ConnConf {
	t.Helper()

	return FlenvRootConnConfStorageTyped(t, storage.PostgreSQLStorageType)
}
