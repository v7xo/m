package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

var (
	ResourceOpenOKCount = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("resource"),
		"open_success_count",
		"indicates the count of succeeded open operation of resource (DB conn, file, etc)",
		[]string{
			attrname.ResourceKind.String(),
			attrname.ObjType.String(),
			attrname.ObjName.String(),
		},
	)

	ResourceCloseOKCount = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("resource"),
		"close_success_count",
		"indicates the count of succeeded close operation of resource (DB conn, file, etc)",
		[]string{
			attrname.ResourceKind.String(),
			attrname.ObjType.String(),
			attrname.ObjName.String(),
		},
	)
)
