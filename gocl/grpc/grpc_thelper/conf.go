package grpc_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/grpc/client"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/server"
)

func FlenvRootServerConf(t *testing.T) server.Conf {
	t.Helper()

	c, err := server.FlenvRootConf()
	require.NoError(t, err, "unexpected error while getting gRPC server configuration")

	return c
}

func FlenvRootClientConf(t *testing.T) client.Conf {
	t.Helper()

	c, err := client.FlenvRootConf()
	require.NoError(t, err, "unexpected error while getting gRPC server configuration")

	return c
}
