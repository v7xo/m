package grpc

import (
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
)

const (
	ServerConfFlenvSelfFlagPrefix = "grpc"
	ClientConfFlenvSelfFlagPrefix = "grpc-client"

	DefaultHttpGatewayEnabled         = false
	DefaultHttpGatewayListenHost      = httpext.DefaultListenHost
	DefaultHttpGatewayListenPort      = httpext.DefaultListenPort
	DefaultHttpGatewayShutdownTimeout = httpext.DefaultServerShutdownTimeout
	DefaultListenHost                 = "0.0.0.0"
	DefaultListenPort                 = 8081
	DefaultMaxRecvMsgSize             = ioext.DefaultApplicationBufferSize
	DefaultNoAccessLog                = false
	DefaultNoBodyInAccessLog          = false
	DefaultNoMetrics                  = false
	DefaultPreventErrDetailedMessage  = true

	GRPCAPICallStyle = "grpc"
	StreamCallType   = "stream"
	UnaryCallType    = "unary"
)

type LogVerbosityLevel int // 1-99

const MaxLogVerbosityLevel = LogVerbosityLevel(99)

type LogSeverityLevel string // info, warning, error

const (
	InfoLogSeverityLevel    = LogSeverityLevel("info")
	WarningLogSeverityLevel = LogSeverityLevel("warning")
	ErrorLogSeverityLevel   = LogSeverityLevel("error")
)
