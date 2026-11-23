package server

import (
	"fmt"
	"slices"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type OptionFn func(*impl) error

func (o OptionFn) Host(v string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.grpcListenHost = v })
}

func (o OptionFn) Port(v int) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.listenPort = v })
}

func (o OptionFn) MaxRecvMsgSize(v int64) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.maxRecvMsgSize = v })
}

func (o OptionFn) Conf(v Conf) OptionFn {
	return fopt.Chain(o, func(t *impl) { applyConf(t, v) })
}

func (o OptionFn) GRPCOption(val grpc.ServerOption) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.grpcOpts = append(t.grpcOpts, val) })
}

func (o OptionFn) GRPCOptions(val []grpc.ServerOption) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.grpcOpts = val })
}

func (o OptionFn) GRPCOptionsAppend(val ...grpc.ServerOption) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.grpcOpts = append(t.grpcOpts, val...) })
}

func (o OptionFn) NoReflection() OptionFn {
	return fopt.Chain(o, func(t *impl) { t.noReflection = true })
}

func (o OptionFn) ReflectionProtoFiles(v []protoreflect.FileDescriptor) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.reflectionProtoFiles = v })
}

func (o OptionFn) NoMetrics() OptionFn {
	return fopt.Chain(o, func(t *impl) { t.noMetrics = true })
}

func (o OptionFn) Log(val logext.Logger) OptionFn {
	return fopt.WithLogScope(o, val, LogScope)
}

func (o OptionFn) PreventErrDetailedMessage(v bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.preventErrDetailedMessage = v })
}

func (o OptionFn) Services(val ...clgrpc.ServiceDescGetter) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.services = slices.Clone(val) })
}

func (o OptionFn) ServicesAppend(val ...clgrpc.ServiceDescGetter) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.services = append(t.services, val...) })
}

// Deprecated: use [OptionFn.Services] instead.
func (o OptionFn) Implementation(val any) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			switch typed := val.(type) {
			case nil:
				t.services = []clgrpc.ServiceDescGetter{nil}

				return nil
			case clgrpc.ServiceDescGetter:
				t.services = []clgrpc.ServiceDescGetter{typed}

				return nil
			default:
				return fmt.Errorf("type '%s' of service implementation not implements required type '%s'",
					typestring.TypeStringOfVal(val),
					typestring.TypeStringFor[clgrpc.ServiceDescGetter]())
			}
		})
}

func (o OptionFn) UnaryInterceptors(val ...grpc.UnaryServerInterceptor) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.unaryInterceptors = slices.Clone(val) })
}

func (o OptionFn) UnaryInterceptorsAppend(val ...grpc.UnaryServerInterceptor) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.unaryInterceptors = append(t.unaryInterceptors, val...) })
}

func applyConf(i *impl, c Conf) {
	i.grpcListenHost = c.ListenHost
	i.listenPort = c.ListenPort
	i.maxRecvMsgSize = c.MaxRecvMsgSize
	i.noAccessLog = c.NoAccessLog
	i.noBodyInAccessLog = c.NoBodyInAccessLog
	i.noMetrics = c.NoMetrics
	i.preventErrDetailedMessage = c.PreventErrDetailedMessage
}

func Option() OptionFn { return fopt.Nop[impl] }
