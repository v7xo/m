package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

//nolint:gochecknoglobals
var (
	StorageAPIQueryExecStatsScanSeconds = prommetric.MustNewHistogramVec(
		true,
		prommetric.CLMetricsSubsystemName("storage_api"),
		"query_exec_stats_scan_seconds",
		"aggregates durations of storage API query scan ops in seconds",
		[]string{
			attrname.StorageType.String(),
			attrname.ObjName.String(),
			attrname.StorageQueryName.String(),
			attrname.Status.String(),
		},
	)

	StorageAPIQueryExecStatsExecSeconds = prommetric.MustNewHistogramVec(
		true,
		prommetric.CLMetricsSubsystemName("storage_api"),
		"query_exec_stats_exec_seconds",
		"aggregates durations of storage API query exec ops in seconds",
		[]string{
			attrname.StorageType.String(),
			attrname.ObjName.String(),
			attrname.StorageQueryName.String(),
			attrname.Status.String(),
		},
	)

	StorageAPIQueryExecStatsAcquireConnSeconds = prommetric.MustNewHistogramVec(
		true,
		prommetric.CLMetricsSubsystemName("storage_api"),
		"query_exec_stats_acquire_conn_seconds",
		"aggregates durations of storage API connection acquisition ops in seconds",
		[]string{
			attrname.StorageType.String(),
			attrname.ObjName.String(),
			attrname.Status.String(),
		},
	)

	StorageAPIQueryExecStatsTotalSeconds = prommetric.MustNewHistogramVec(
		true,
		prommetric.CLMetricsSubsystemName("storage_api"),
		"query_exec_stats_total_seconds",
		"aggregates total durations of storage API query exec ops in seconds (time include conn acquire, exec, scan)",
		[]string{
			attrname.StorageType.String(),
			attrname.ObjName.String(),
			attrname.StorageQueryName.String(),
			attrname.Status.String(),
		},
	)

	StorageAPIQueryExecStatsRowsAffectedGaugeVec = prommetric.MustNewGaugeVec(
		true,
		prommetric.CLMetricsSubsystemName("storage_api"),
		"query_exec_stats_rows_affected_count",
		"counts storage API rows affected by query",
		[]string{
			attrname.StorageType.String(),
			attrname.ObjName.String(),
			attrname.StorageQueryName.String(),
			attrname.Status.String(),
		},
	)
)
