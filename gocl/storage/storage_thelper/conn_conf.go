package storage_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

func FlenvRootConnConfStorageTyped(t *testing.T, st storage.StorageType) storage.ConnConf {
	t.Helper()

	conf := FlenvRootConnConf(t)
	require.Equal(t, st.String(), conf.Type, "unexpected storage type specified by default storage connection configuration")

	return conf
}

func FlenvRootConnConf(t *testing.T) storage.ConnConf {
	t.Helper()

	conf, err := storage.FlenvRootConnConf()
	require.NoError(t, err, "unexpected error while getting default DB connection configuration")

	return conf
}

func FlenvRootMasterDBConf(t *testing.T) storage.MasterDBConf {
	t.Helper()

	conf, err := storage.FlenvRootMasterDBConf()
	require.NoError(t, err, "unexpected error while getting default master DB configuration")

	return conf
}
