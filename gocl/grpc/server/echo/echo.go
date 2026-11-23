package echo

import (
	"context"

	"google.golang.org/grpc"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/pbt/types"
)

type echo struct{}

var _ clgrpc.ServiceDescGetter = &echo{}

func (echo) GRPCServiceDesc() (*grpc.ServiceDesc, error) { return &EchoService_ServiceDesc, nil }

func (echo) EchoV1(ctx context.Context, req *types.EchoV1Request) (res *types.EchoV1Response, err error) {
	return &types.EchoV1Response{
		ReplyText:         req.GetText(),
		ReplyClientIdLong: info.UserAgentLong(),
		ReplyBuildInfo:    info.BuildInfoString(),
	}, nil
}

func New() Echo { return &echo{} }
