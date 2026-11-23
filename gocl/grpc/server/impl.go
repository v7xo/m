package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	grpc_mv "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"

	"rn.m.vo7.io/vo7io/m/gocl/app/appmodule"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/mw/contextenrich"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/mw/metrics"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/server"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	clpbreflect "rn.m.vo7.io/vo7io/m/gocl/misc/pbreflect"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	clgrpcstatus "rn.m.vo7.io/vo7io/m/gocl/status/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var (
	_ appmodule.Module
	_ readiness.Collector
)

type impl struct {
	fopt.LogOpt

	grpcSrvLn                 atomic.Pointer[net.Listener]
	grpcSrv                   *grpc.Server
	grpcListenHost            string
	services                  []clgrpc.ServiceDescGetter
	unaryInterceptors         []grpc.UnaryServerInterceptor
	grpcOpts                  []grpc.ServerOption
	reflectionProtoFiles      []protoreflect.FileDescriptor
	listenPort                int
	maxRecvMsgSize            int64
	grpcSrvReadiness          atomic.Bool
	noAccessLog               bool
	noBodyInAccessLog         bool
	noMetrics                 bool
	noReflection              bool
	preventErrDetailedMessage bool
}

// Run implements [appmodule.Module.Name] app module run.
func (*impl) Name() string { return server.AppModuleName }

// Run implements [appmodule.Module.Run] app module run.
func (s *impl) Run(ctx context.Context) error { return s.Serve(ctx) }

// CollectReadinessConsolidated implements [readiness.Collector].
func (s *impl) CollectReadinessConsolidated(_ context.Context) (bool, error) {
	return s.grpcSrvReadiness.Load(), nil
}

func (s *impl) initializeGRPC(ctx context.Context) (err error) {
	grpcOpts := append(defaultGRPCOpts(), s.grpcOpts...)
	grpcOpts = append(grpcOpts, grpc.MaxRecvMsgSize(int(s.maxRecvMsgSize)))

	unaryInterceptorGRPCOpt, err := buildGRPCUnaryInterceptorOpt(s)
	if err != nil {
		return
	} else if unaryInterceptorGRPCOpt != nil {
		grpcOpts = append(grpcOpts, unaryInterceptorGRPCOpt)
	}

	s.grpcSrv = grpc.NewServer(grpcOpts...)

	for i := range len(s.services) {
		desc, descErr := s.services[i].GRPCServiceDesc()
		if descErr != nil {
			return fmt.Errorf("failed to get gRPC service description for service '%s' at '%d' position; %w", typestring.TypeStringOfVal(s.services[i]), i, descErr)
		}

		s.grpcSrv.RegisterService(desc, s.services[i])
		s.Log().Trace(ctx).Msgf("gRPC service registered, service type '%s'", typestring.TypeStringOfVal(s.services[i]))
	}

	if !s.noReflection {
		if err = clpbreflect.RegisterProtoFilesGlobally(s.reflectionProtoFiles); err != nil {
			return
		}

		reflection.Register(s.grpcSrv)
		s.Log().Trace(ctx).Msg("reflection service registered")
	}

	if !s.noMetrics {
		grpc_prom.EnableHandlingTimeHistogram()

		if _, err = grpc_prom.RegisterDefaultServerMetricsWithRegisterer(prommetric.DefaultRegistrer()); err != nil {
			return
		}

		grpc_prom.Register(s.grpcSrv)
	}

	return nil
}

// TODO: make as a separate mw.
func makeAccessLogInterceptor(log logext.Logger, omitBody bool) grpc.UnaryServerInterceptor {
	var (
		reqRespBodyEnricher func(e logext.Event, req any, resp any) logext.Event
		reqMDEnriched       func(e logext.Event, md metadata.MD) logext.Event
	)

	if log.LogLevel().IsTraceOrDebug() {
		reqMDEnriched = logEventEnrichWithGRPCIncomingMD
	} else {
		reqMDEnriched = func(e logext.Event, _ metadata.MD) logext.Event { return e }
	}

	if log.LogLevel().IsTraceOrDebug() && !omitBody {
		reqRespBodyEnricher = logEventEnrichWithReqRespBody
	} else {
		reqRespBodyEnricher = func(e logext.Event, _, _ any) logext.Event { return e }
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		reqMd, _ := metadata.FromIncomingContext(ctx)
		start := time.Now()
		resp, err = handler(ctx, req)
		dur := time.Since(start)

		var logEvent logext.Event

		if err == nil {
			logEvent = log.
				Info(ctx).
				Int64(attrname.TimeDurationMs, dur.Milliseconds()).
				Uint32(attrname.StatusCode, uint32(codes.OK)).
				Str(attrname.Status, string(opstatus.OK))
		} else {
			logEvent = log.LevelFromErr(ctx, err)
			errStatus, _ := errs.TranslateGRPCStatus(err)

			logEvent = logEvent.
				Int64(attrname.TimeDurationMs, dur.Milliseconds()).
				Uint32(attrname.StatusCode, uint32(errStatus.Code())).
				Str(attrname.Status, string(errs.TranslateStatus(err))).
				Str(attrname.StatusCodeName, errStatus.Code().String())
		}

		logEvent = reqMDEnriched(logEvent, reqMd)
		logEvent = reqRespBodyEnricher(logEvent, req, resp)

		logEvent.
			Bool(attrname.AccessLog, true).
			Str(attrname.RPCMethod, info.FullMethod).
			Str(attrname.RPCType, clgrpc.UnaryCallType).
			Send()

		return
	}
}

// TODO: make as a separate mw.
func makeErrStatusTranslatorInterceptor(preventErrDetailedMessage bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, hnd grpc.UnaryHandler) (resp any, err error) {
		if resp, err = hnd(ctx, req); err != nil {
			if errStatus, exactStatus := clgrpcstatus.Translate(err); exactStatus {
				switch {
				case preventErrDetailedMessage:
					err = errStatus.Err()
				default:
					err = clgrpcstatus.ChangeMessage(errStatus, err.Error()).Err()
				}
			} else if preventErrDetailedMessage {
				err = status.Error(errStatus.Code(), placeholder.ErrDetailedMessageOmitted)
			}
		}

		return
	}
}

// TODO: make as a separate mw.
func makePanicInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, hnd grpc.UnaryHandler) (resp any, err error) {
		panicking := true

		defer func() {
			if !panicking {
				return
			}

			err = errors.Join(err, errs.PanicErrUnconditionally(recover()))
		}()

		resp, err = hnd(ctx, req)
		panicking = false

		return
	}
}

func createListener(stopCtx context.Context, host string, port int) (net.Listener, error) {
	lnCfg := net.ListenConfig{}

	lnAddr := net.JoinHostPort(host, strconv.Itoa(port))

	ln, err := lnCfg.Listen(stopCtx, "tcp", lnAddr)
	if err != nil {
		err = fmt.Errorf("error while announcing local network address; addr: %s; %w", lnAddr, err)

		return nil, err
	}

	return ln, nil
}

func buildGRPCUnaryInterceptor(srv *impl, log logext.Logger) (grpc.UnaryServerInterceptor, error) {
	var interceptors []grpc.UnaryServerInterceptor

	if !srv.noMetrics {
		interceptors = append(interceptors, grpc_prom.UnaryServerInterceptor)
	}

	interceptors = append(interceptors, makeErrStatusTranslatorInterceptor(srv.preventErrDetailedMessage))

	ctxEnrich, err := contextenrich.New(contextenrich.Option().Log(log))
	if err != nil {
		return nil, err
	}

	interceptors = append(interceptors, ctxEnrich)

	if !srv.noMetrics {
		// TODO: in future versions mw below should replace third-party [grpc_prom.UnaryServerInterceptor].
		var metricsUnary grpc.UnaryServerInterceptor

		if metricsUnary, err = metrics.New(metrics.Option().Log(log)); err != nil {
			return nil, err
		}

		interceptors = append(interceptors, metricsUnary)
	}

	if !srv.noAccessLog && log != nil && !log.IsNop() {
		interceptors = append(interceptors, makeAccessLogInterceptor(log, srv.noBodyInAccessLog))
	}

	interceptors = append(interceptors, makePanicInterceptor())
	interceptors = append(interceptors, srv.unaryInterceptors...)

	switch len(interceptors) {
	case 0:
		return nil, nil
	case 1:
		return interceptors[0], nil
	default:
		return grpc_mv.ChainUnaryServer(interceptors...), nil
	}
}

func buildGRPCUnaryInterceptorOpt(srv *impl) (grpc.ServerOption, error) {
	mw, err := buildGRPCUnaryInterceptor(srv, srv.Log())
	if err != nil {
		return nil, err
	}

	return grpc.UnaryInterceptor(mw), nil
}

// TODO: make it configurable via Config object.
func defaultGRPCOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveParams(DefaultKeepAliveParams()),
		grpc.MaxConcurrentStreams(DefaultMaxConcurrentStreams),
		grpc.Creds(insecure.NewCredentials()),
	}
}

func New(opts ...OptionFn) (_ Server, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if len(impl.services) == 0 {
		return nil, errors.New("no any gRPC service specified")
	}

	for i := range len(impl.services) {
		if impl.services[i] == nil {
			return nil, fmt.Errorf("nil gRPC service specified at '%d' position", i)
		}
	}

	if err = impl.SetLog(impl.Log().Tune().NextHook(impl.logEventEnrichGRPC).Tuned()); err != nil {
		return nil, err
	}

	return impl, nil
}
