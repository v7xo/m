package contextenrich

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/md"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type impl struct {
	fopt.LogOpt

	enabled bool
}

func (s *impl) unary(ctx context.Context, req any, opInfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if !s.enabled {
		return handler(ctx, req)
	}

	logAttrs := make(attrname.Dict, 5) //nolint

	switch incomingMD, incomingMDExists := metadata.FromIncomingContext(ctx); incomingMDExists {
	case true:
		if val, ok := md.ExtractLastString(incomingMD, attrname.XRequestID); ok {
			logAttrs[attrname.RequestID] = val
			ctx = contextext.WithRequestIDReplace(ctx, val)
		}

		if val, ok := md.ExtractLastString(incomingMD, attrname.XUserAgent); ok {
			logAttrs[attrname.UserAgent] = val
		}

		if val, ok := md.ExtractLastString(incomingMD, attrname.XUserAgentAppCmd); ok {
			logAttrs[attrname.UserAgentAppCmd] = val
		}
	}

	logAttrs[attrname.RPCMethod] = opInfo.FullMethod
	logAttrs[attrname.RPCType] = clgrpc.UnaryCallType

	ctx = attrname.DictContext(ctx, logAttrs)

	return handler(ctx, req)
}

func New(opts ...OptionFn) (grpc.UnaryServerInterceptor, error) {
	const errMsgTemplate = "failed to create instance of gRPC context enrichment unary interceptor (middleware); err: %w"

	res := &impl{
		enabled: true,
	}

	err := fopt.Apply(res, opts...)
	if err != nil {
		return nil, fmt.Errorf(errMsgTemplate, err)
	}

	return res.unary, nil
}
