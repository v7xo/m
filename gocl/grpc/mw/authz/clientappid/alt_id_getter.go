package clientappid

import (
	"context"

	"google.golang.org/grpc"
)

type AltGetter func(ctx context.Context, req any, info *grpc.UnaryServerInfo) (string, bool, error)
