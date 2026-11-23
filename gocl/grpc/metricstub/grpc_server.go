package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

var (
	GRPCServerClientAppHandledTotal = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("grpc_server"),
		"client_app_handled_total",
		"indicates the count of handled gRPC requests",
		[]string{
			string(attrname.RPCService),
			string(attrname.RPCMethod),
			string(attrname.RPCType),
			string(attrname.ClientAppID),
			string(attrname.StatusCode),
			string(attrname.UserAgent),
			string(attrname.UserAgentAppCmd),
			string(attrname.UserIdentityProvider),
			string(attrname.UserRole),
		})

	GRPCServerClientAppHandledSeconds = prommetric.MustNewHistogramVec(
		false,
		prommetric.CLMetricsSubsystemName("grpc_server"),
		"client_app_handled_seconds",
		"indicates the count/duration of handled gRPC requests",
		[]string{
			string(attrname.RPCService),
			string(attrname.RPCMethod),
			string(attrname.RPCType),
			string(attrname.StatusCode),
			string(attrname.UserAgent),
			string(attrname.UserAgentAppCmd),
			string(attrname.UserIdentityProvider),
			string(attrname.UserRole),
		})
)
