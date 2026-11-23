package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	mdpkg "rn.m.vo7.io/vo7io/m/gocl/grpc/md"
	"rn.m.vo7.io/vo7io/m/gocl/grpc/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	enabled bool
}

func (*impl) serviceMethodName(fullMethodName string) (serviceName string, methodName string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")

	if before, after, ok := strings.Cut(fullMethodName, "/"); ok {
		serviceName, methodName = before, after

		return
	}

	return
}

func (s *impl) unary(ctx context.Context, req any, opInfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !s.enabled {
		resp, err = handler(ctx, req)

		return
	}

	startTs := time.Now().UTC()

	userIdentityProvider := placeholder.Unknown
	userAgent := placeholder.Unknown
	userAgentAppCmd := placeholder.Unknown
	userRole := placeholder.Unknown

	switch incomingMD, incomingMDExists := metadata.FromIncomingContext(ctx); incomingMDExists {
	case true:
		if val, ok := mdpkg.ExtractLastString(incomingMD, attrname.XUserAgent); ok {
			userAgent = val
		}

		if val, ok := mdpkg.ExtractLastString(incomingMD, attrname.XUserAgentAppCmd); ok {
			userAgentAppCmd = val
		}
	}

	serviceName, methodName := s.serviceMethodName(opInfo.FullMethod)
	serviceName, methodName = placeholder.UnknownIfEmpty(serviceName), placeholder.UnknownIfEmpty(methodName)

	defer func() {
		dur := time.Since(startTs).Seconds()

		metricstub.
			GRPCServerClientAppHandledTotal.
			WithLabelVals(
				serviceName,
				methodName,
				clgrpc.UnaryCallType,
				placeholder.Omitted, // label [attrname.ClientAppID] deprecated in context of metric.
				string(errs.TranslateStatus(err)),
				userAgent,
				userAgentAppCmd,
				userIdentityProvider,
				userRole,
			).
			Inc()

		metricstub.
			GRPCServerClientAppHandledSeconds.
			WithLabelVals(
				serviceName,
				methodName,
				clgrpc.UnaryCallType,
				string(errs.TranslateStatus(err)),
				userAgent,
				userAgentAppCmd,
				userIdentityProvider,
				userRole,
			).
			Observe(dur)
	}()

	resp, err = handler(ctx, req)

	return
}

func New(opts ...OptionFn) (grpc.UnaryServerInterceptor, error) {
	const errMsgTemplate = "failed to create instance of gRPC server metrics unary interceptor (middleware); err: %w"

	res := &impl{
		enabled: true,
	}

	err := fopt.Apply(res, opts...)
	if err != nil {
		return nil, fmt.Errorf(errMsgTemplate, err)
	}

	return res.unary, nil
}
