package grpc_thelper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/client"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/client/echo"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/server"
	"rn.m.vo7.io/vo7io/m/gocl/misc/netext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
	pbt "rn.m.vo7.io/vo7io/m/gocl/pbt/types"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

type ServeOptionFn[TClient echo.Echo] func(*srvbldr[TClient])

func (o ServeOptionFn[TClient]) ClientConf(val func(t *testing.T) client.Conf) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.clientConf = val }
}

func (o ServeOptionFn[TClient]) ClientConfEdit(val func(*testing.T, *client.Conf)) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.clientConfEdit = val }
}

func (o ServeOptionFn[TClient]) ClientOptionEdit(val func(*testing.T, client.OptionFn[TClient]) client.OptionFn[TClient]) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.clientOptEdit = val }
}

func (o ServeOptionFn[TClient]) ClientConstructor(val func(cc grpc.ClientConnInterface) TClient) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.clientConstructor = val }
}

func (o ServeOptionFn[TClient]) ServerConf(val func(t *testing.T) server.Conf) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.serverConf = val }
}

func (o ServeOptionFn[TClient]) ServerConfEdit(val func(*testing.T, *server.Conf)) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.serverConfEdit = val }
}

func (o ServeOptionFn[TClient]) ServerOptionEdit(val func(*testing.T, server.OptionFn) server.OptionFn) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.serverOptEdit = val }
}

func (o ServeOptionFn[TClient]) ServiceImplementation(val func(*testing.T) any) ServeOptionFn[TClient] {
	return func(s *srvbldr[TClient]) { o(s); s.serviceImplementation = val }
}

func ServeOption[TClient echo.Echo]() ServeOptionFn[TClient] {
	return func(f *srvbldr[TClient]) {}
}

type srvbldr[TClient echo.Echo] struct {
	clientConf            func(t *testing.T) client.Conf
	clientConfEdit        func(*testing.T, *client.Conf)
	clientConstructor     func(cc grpc.ClientConnInterface) TClient
	clientOptEdit         func(*testing.T, client.OptionFn[TClient]) client.OptionFn[TClient]
	serverConf            func(t *testing.T) server.Conf
	serverConfEdit        func(*testing.T, *server.Conf)
	serverOptEdit         func(*testing.T, server.OptionFn) server.OptionFn
	serviceImplementation func(*testing.T) any
}

func newSrvbldr[TClient echo.Echo](t *testing.T, opts ...ServeOptionFn[TClient]) *srvbldr[TClient] {
	t.Helper()

	bldr := &srvbldr[TClient]{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(bldr)
	}

	if bldr.serverConf == nil {
		bldr.serverConf = func(t *testing.T) server.Conf {
			t.Helper()

			return FlenvRootServerConf(t)
		}
	}

	if bldr.clientConf == nil {
		bldr.clientConf = func(t *testing.T) client.Conf {
			t.Helper()

			return FlenvRootClientConf(t)
		}
	}

	require.NotNil(t, bldr.serviceImplementation, "no service implementation getter function specified")
	require.NotNil(t, bldr.clientConstructor, "no gRPC client constructor function specified")

	return bldr
}

func Serve[TClient echo.Echo](t *testing.T, ctx context.Context, opts ...ServeOptionFn[TClient]) (serverDead context.Context, client TClient) {
	t.Helper()

	bldr := newSrvbldr(t, opts...)

	serverConf := bldr.serverConf(t)
	if bldr.serverConfEdit != nil {
		bldr.serverConfEdit(t, &serverConf)
	}

	require.Truef(t, serverConf.NoMetrics, "gRPC metrics emitting should be disabled on server configuratiion level because of not supported by this test helper")

	implementation := bldr.serviceImplementation(t)
	require.NotNil(t, implementation, "unexpected nil service implementation")

	log := logext_thelper.FromContext(t, ctx)
	serverOpt := server.Option().
		Conf(serverConf).
		Implementation(implementation).
		Log(log)

	if bldr.serverOptEdit != nil {
		serverOpt = bldr.serverOptEdit(t, serverOpt)
	}

	server, err := server.New(serverOpt)
	require.NoError(t, err)
	require.NotNil(t, server)

	grpcLn, portNum := netext_thelper.RandomTCP4Listen(t, ctx)
	stop, stopSig := context.WithCancel(ctx)
	serverDead, serverDeadSig := context.WithCancel(stop)
	serveErrCh := make(chan error, 1)

	// serve gRPC.
	go func() {
		defer serverDeadSig()

		serveErrCh <- server.ServeWithListener(grpcLn, stop)
	}()

	go func() {
		if err := <-serveErrCh; err != nil {
			log.Err(ctx, err).Msg("error while serving gRPC")

			if !t.Failed() {
				defer func() {
					if panicErr := errs.PanicErr(recover()); panicErr != nil {
						log.Err(ctx, err).Msgf("error while marking test '%s' as failed", t.Name())
					}
				}()

				t.Fail()
			}
		}

		close(serveErrCh)
	}()

	thelper.Cleanup(
		t,
		func() error {
			stopSig()

			select {
			case <-serveErrCh: // nolint:errcheck
			case <-time.After(time.Second * 5): // nolint:mnd,revive
				return fmt.Errorf("gRPC server not responded correctly on stop signal during allotted time '%s'", time.Second*time.Duration(5)) // nolint:mnd,revive
			}

			return nil
		},
	)

	client = dialAndEchoClient(t, serverDead, portNum, bldr)

	return
}

func dialAndEchoClient[TClient echo.Echo](t *testing.T, ctx context.Context, portNum int, bldr *srvbldr[TClient]) TClient {
	t.Helper()

	clientConf := bldr.clientConf(t)
	if bldr.clientConfEdit != nil {
		bldr.clientConfEdit(t, &clientConf)
	}

	clientConf.Port = portNum
	clientConf.Host = ""

	log := logext_thelper.FromContext(t, ctx)

	clientOpt := client.Option[TClient]().
		Conf(clientConf).
		ClientConstructor(bldr.clientConstructor).
		Log(log)

	if bldr.clientOptEdit != nil {
		clientOpt = bldr.clientOptEdit(t, clientOpt)
	}

	cli, cliCloser, err := client.New(clientOpt)
	require.NoError(t, err)
	require.NotNil(t, cli, "unexpected nil gRPC client")
	require.NotNil(t, cliCloser, "unexpected nil gRPC client connection close function")

	t.Cleanup(cliCloser)

	randomText := uuid_thelper.Rand4BHex(t)
	resp, err := cli.EchoV1(ctx, &pbt.EchoV1Request{Text: randomText})
	require.NoError(t, err, "unexpected error result of Echo request to gRPC service")
	require.NotNil(t, resp, "unexpected nil-result of Echo request to gRPC service")
	require.Equal(t, randomText, resp.GetReplyText(), "gRPC service response text not equals to expected one")

	log.Info(ctx).Msg("gRPC service/client: 'Echo' check passed")

	return cli
}
