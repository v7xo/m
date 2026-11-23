package storageapi_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

func FlenvRootConf(t *testing.T) storageapi.Conf {
	t.Helper()

	c, err := storageapi.FlenvRootConf()
	require.NoError(t, err, "unexpected error while getting storage API configuration")

	return c
}

func FlenvRootClientConf(t *testing.T) storageapi.ClientConf {
	t.Helper()

	c, err := storageapi.FlenvRootClientConf()
	require.NoError(t, err, "unexpected error while getting storage API connection configuration")

	return c
}
