package server

import (
	"time"

	"google.golang.org/grpc/keepalive"

	"rn.m.vo7.io/vo7io/m/gocl/grpc"
)

const (
	// AppModuleName denotes name of HTTP server as application module.
	AppModuleName = "grpc_server"

	DefaultMaxConcurrentStreams = 64
	ConfFlenvSelfFlagPrefix     = grpc.ServerConfFlenvSelfFlagPrefix

	LogScope = "gocl_grpc_server"
)

//nolint:revive,mnd
func DefaultKeepAliveParams() keepalive.ServerParameters {
	return keepalive.ServerParameters{
		Time:    time.Minute * 5,
		Timeout: time.Second * 5,
	}
}
