package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"

	grpc "google.golang.org/grpc"
	creds "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/client/echo"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type (
	ClientConstructorFn[TClient echo.Echo] func(cc grpc.ClientConnInterface) TClient
	ClientConnCloseFn                      func()

	Factory[TClient echo.Echo] interface {
		isGRPCClientFactory()

		New() (TClient, ClientConnCloseFn, error)
	}

	factory[TClient echo.Echo] struct {
		fopt.LogOpt
		clientConstructor ClientConstructorFn[TClient]
		conf              Conf
	}
)

func (s *factory[TClient]) New() (c TClient, closeConn ClientConnCloseFn, err error) {
	conn, closeConn, err := s.makeClientConn()
	if err != nil {
		return
	}

	c = s.clientConstructor(conn)

	return
}

func (s *factory[TClient]) makeTransportCredentials() (res creds.TransportCredentials, err error) {
	switch {
	case s.conf.TransportCredentialsPreset().IsPlaintext():
		return insecure.NewCredentials(), nil
	case s.conf.TransportCredentialsPreset().IsTLSInsecure():
		// nolint:gosec
		res = creds.NewTLS(
			&tls.Config{
				InsecureSkipVerify: true,
				ClientAuth:         tls.NoClientCert,
			},
		)

		return
	case s.conf.TransportCredentialsPreset().IsTLSSecureClient():
		if len(s.conf.ClientSSLCertPEM) != 0 && len(s.conf.ClientSSLKeyPEM) != 0 {
			var cert tls.Certificate

			cert, err = tls.X509KeyPair([]byte(s.conf.ClientSSLCertPEM), []byte(s.conf.ClientSSLKeyPEM))
			if err != nil {
				err = fmt.Errorf("error while constructing X.509 key pair for specified client SSL certificate & key; %w", err)

				return nil, err
			}

			res = creds.NewTLS(
				&tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					MinVersion:   tls.VersionTLS13,
				},
			)
		}

		return
	case s.conf.TransportCredentialsPreset().IsTLSSecureServer():
		err = fmt.Errorf("gRPC transport credentials for server are not supported; %w", errs.ErrNotSupported)

		return
	default:
		err = fmt.Errorf("can't determine gRPC transport credentials from specified configuration"+
			" (either plaintext or tls-insecure option should be specified by configuration); %w", errs.ErrNotSupported)

		return
	}
}

func (s *factory[TClient]) requestIDUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	outgoingMD, outgoingMDExists := metadata.FromOutgoingContext(ctx)
	existing := outgoingMD.Get(string(attrname.XRequestID))

	if outgoingMDExists && len(existing) != 0 && len(existing[len(existing)-1]) != 0 {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	reqID, reqIDExists, err := contextext.RequestID(ctx)
	if err != nil {
		return err
	} else if !reqIDExists {
		switch reqID, err = uuidext.RequestIDHex(); err != nil {
		case true:
			return err
		}

		s.Log().
			Info(ctx).
			Str(attrname.RequestID, reqID).
			Str(attrname.RPCMethod, method).
			Msgf("use newly generated request ID '%s' for this gRPC call", reqID)
	} else {
		s.Log().
			Info(ctx).
			Str(attrname.RequestID, reqID).
			Str(attrname.RPCMethod, method).
			Msgf("use context provided request ID '%s' for this gRPC call", reqID)
	}

	ctx = metadata.AppendToOutgoingContext(ctx, string(attrname.XRequestID), reqID)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func (s *factory[TClient]) errorEnricherUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		return fmt.Errorf("failed to make unary call to gRPC method '%s'; %w", method, err)
	}

	return nil
}

func (s *factory[TClient]) userAgentUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		string(attrname.XUserAgent), info.UserAgentShort(),
		string(attrname.XUserAgentAppCmd), info.CmdAppQualifiedName(),
	)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func (s *factory[TClient]) makeDialOpts() ([]grpc.DialOption, error) {
	transportCreds, err := s.makeTransportCredentials()
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithChainUnaryInterceptor(
			grpc.UnaryClientInterceptor(s.errorEnricherUnaryInterceptor),
			grpc.UnaryClientInterceptor(s.requestIDUnaryInterceptor),
			grpc.UnaryClientInterceptor(s.userAgentUnaryInterceptor),
		),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(int(s.conf.MaxRecvMsgSize)),
		),
	}

	return opts, nil
}

func (s *factory[TClient]) makeClientConn() (conn grpc.ClientConnInterface, closeConn ClientConnCloseFn, err error) {
	dialOpts, err := s.makeDialOpts()
	if err != nil {
		return
	}

	addr := net.JoinHostPort(s.conf.Host, strconv.Itoa(s.conf.Port))

	client, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		err = fmt.Errorf("failed to create gRPC client; addr: %s; %w", addr, err)

		return
	}

	closeConn = func() {
		if closeErr := client.Close(); closeErr != nil {
			s.Log().
				Err(context.Background(), closeErr).
				Str(attrname.AddrHost, s.conf.Host).
				Int(attrname.AddrPort, s.conf.Port).
				Msg("error while closing this gRPC client connection")
		}
	}

	conn = client

	return
}

// isGRPCClientFactory implements [Factory`1] interface.
func (s *factory[TClient]) isGRPCClientFactory() {} // nolint

func New[TClient echo.Echo](opts ...OptionFn[TClient]) (_ TClient, _ ClientConnCloseFn, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[factory[TClient]](), err)
		}
	}()

	var (
		zero    TClient
		factory = &factory[TClient]{}
	)

	if err = fopt.Apply(factory, opts...); err != nil {
		return zero, nil, err
	}

	if factory.clientConstructor == nil {
		return zero, nil, errors.New("gRPC client constructor function not specified")
	}

	return factory.New()
}

func NewFactory[TClient echo.Echo](opts ...OptionFn[TClient]) (res Factory[TClient], err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[factory[TClient]](), err)
		}
	}()

	factory := &factory[TClient]{}

	if err = fopt.Apply(factory, opts...); err != nil {
		return
	}

	if factory.clientConstructor == nil {
		return nil, errors.New("gRPC client constructor function not specified")
	}

	return factory, nil
}

func NewFactoryConf[TClient echo.Echo](conf Conf, constructor ClientConstructorFn[TClient], log logext.Logger) (Factory[TClient], error) {
	return NewFactory(
		Option[TClient]().
			ClientConstructor(constructor).
			Conf(conf).
			Log(log),
	)
}
