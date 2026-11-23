package server

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"rn.m.vo7.io/vo7io/m/gocl/app/appmodule"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
)

var _ grpc.Server

// Server is wrapper-interface for [grpc.Server].
type Server interface {
	appmodule.Module
	readiness.Collector
	// Serve starts serving incoming requests for this server.
	Serve(stop context.Context) error
	// ServeWithListener starts serving incoming requests for this server.
	ServeWithListener(ln net.Listener, stop context.Context) error
}
