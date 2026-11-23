package grpc

import "google.golang.org/grpc"

type ServiceDescGetter interface {
	GRPCServiceDesc() (*grpc.ServiceDesc, error)
}
