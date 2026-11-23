package echo

import (
	"context"

	"google.golang.org/grpc"

	"rn.m.vo7.io/vo7io/m/gocl/pbt/types"
)

type Echo interface {
	EchoV1(ctx context.Context, req *types.EchoV1Request, opts ...grpc.CallOption) (*types.EchoV1Response, error)
}
