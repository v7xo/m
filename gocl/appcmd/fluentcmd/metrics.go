package fluentcmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

//nolint:gochecknoglobals
var (
	appCmdInfoGaugeVec = promauto.
				With(prommetric.DefaultRegistrer()).
				NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: prommetric.MetricsNamespaceName(true),
				Subsystem: prommetric.CLMetricsSubsystemName("app"),
				Name:      "cmd_info",
				Help:      "build info about running Go app command",
			},
			[]string{
				attrname.AppName.String(),
				attrname.AppVersion.String(),
				attrname.AppCmdName.String(),
			})

	appCmdExitStatusGaugeVec = promauto.
					With(prommetric.DefaultRegistrer()).
					NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: prommetric.MetricsNamespaceName(true),
				Subsystem: prommetric.CLMetricsSubsystemName("app"),
				Name:      "cmd_exit_status",
				Help:      "indicates Go app command exit status",
			},
			[]string{
				attrname.AppName.String(),
				attrname.AppVersion.String(),
				attrname.AppCmdName.String(),
				attrname.Status.String(),
			})
)

func registerAppCmdInfo() {
	appCmdInfoGaugeVec.
		WithLabelValues(info.Name(), info.Version().String(), info.CmdName()).
		Set(1)
}

func registerAppCmdExitStatus(cmdErr error) {
	appCmdExitStatusGaugeVec.
		WithLabelValues(
			info.Name(),
			info.Version().String(),
			info.CmdName(),
			errs.TranslateStatus(cmdErr).String()).
		Inc()
}
