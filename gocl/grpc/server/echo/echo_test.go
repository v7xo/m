//go:build test_unit

package echo_test

import (
	"testing"

	grpc "google.golang.org/grpc"

	echo_client "rn.m.vo7.io/vo7io/m/gocl/grpc/client/echo"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/grpc_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/server"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/server/echo"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_EchoServe(t *testing.T) {
	ctx := thelper.Pre(t)

	serverDead, client := grpc_thelper.Serve(
		t,
		ctx,
		grpc_thelper.ServeOption[echo_client.Echo]().
			ClientConstructor(func(cc grpc.ClientConnInterface) echo_client.Echo { return echo.NewEchoServiceClient(cc) }).
			ServerConfEdit(func(t *testing.T, c *server.Conf) { c.NoMetrics = true }).
			ServiceImplementation(func(t *testing.T) any { return echo.New() }),
	)
	_, _ = serverDead, client
}
