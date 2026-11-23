package prommetric

import (
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/buildinfo"
)

// NormalizeMetricName replaces any invalid symbol in Prometheus metric name n with _ symbol.
// see github.com/prometheus/common/model/MetricNameRE - regexp for metric name validation.
func NormalizeMetricName(n string) string {
	if len(n) == 0 {
		return n
	}

	sb := strings.Builder{}
	sb.Grow(len(n))

	for i, b := range n {
		//nolint:staticcheck
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == ':' || (b >= '0' && b <= '9' && i > 0)) {
			sb.WriteRune('_')

			continue
		}

		sb.WriteRune(b)
	}

	return sb.String()
}

func MetricsNamespaceName(normalized bool) string {
	if buildinfo.MetricsNamespace == "" {
		return buildinfo.MetricsNamespaceFallback
	}

	if normalized {
		return NormalizeMetricName(buildinfo.MetricsNamespace)
	}

	return buildinfo.MetricsNamespace
}

func CLMetricsSubsystemName(component string) string {
	return NormalizeMetricName(CLMetricsSubsystemPrefix + "_" + component)
}
