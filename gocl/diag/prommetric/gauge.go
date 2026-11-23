package prommetric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
)

type GaugeVec struct {
	vec      *prometheus.GaugeVec
	appOwned bool
}

func (m *GaugeVec) WithLabelVals(lv ...string) prometheus.Gauge {
	return m.vec.WithLabelValues(normalizeLabelVal(prependPredefinedLabelVal(lv, !m.appOwned))...)
}

func NewGaugeVec(appOwned bool, component, name, help string, ln []string) (*GaugeVec, error) {
	err := validateUserLabelName(ln)
	if err != nil {
		return nil, err
	}

	var ns string

	ln = prependPredefinedLabelName(ln, !appOwned)

	if appOwned {
		ns = NormalizeMetricName(MetricsNamespaceName(false) + "_" + info.Name())
	} else {
		ns = MetricsNamespaceName(true)
	}

	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Subsystem: NormalizeMetricName(component),
			Name:      NormalizeMetricName(name),
			Help:      help,
		},
		ln)

	if err = DefaultRegistrer().Register(vec); err != nil {
		return nil, err
	}

	return &GaugeVec{vec: vec, appOwned: appOwned}, nil
}

func MustNewGaugeVec(appOwned bool, component, name, help string, labelNames []string) *GaugeVec {
	vec, err := NewGaugeVec(appOwned, component, name, help, labelNames)
	if err != nil {
		panic(fmt.Errorf("unable to create new instance of Gauge metric (%T); err: %w", &GaugeVec{}, err))
	}

	return vec
}
