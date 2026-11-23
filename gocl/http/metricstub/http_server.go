package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

var HTTPServerReqDurSeconds = prommetric.MustNewHistogramVec(
	false,
	prommetric.CLMetricsSubsystemName("http_server"),
	"req_dur_seconds2",
	"indicates the count/duration of HTTP request handled by specified server",
	[]string{
		string(attrname.ServerName),
		string(attrname.HTTPReqMethod),
		string(attrname.HTTPReqURLPath),
		string(attrname.Status),
		string(attrname.StatusCode),
	},
)
