package client

import (
	"rn.m.vo7.io/vo7io/m/gocl/grpc"
)

const (
	DefaultTransportCredentialsPreset = grpc.TransportCredentialsPresetPlaintext
	LogScope                          = "grpc_client"
)
