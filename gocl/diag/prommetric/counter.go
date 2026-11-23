package prommetric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
)

type CounterVec struct {
	vec      *prometheus.CounterVec
	appOwned bool
}

func (m *CounterVec) WithLabelVals(lv ...string) prometheus.Counter {
	return m.vec.WithLabelValues(normalizeLabelVal(prependPredefinedLabelVal(lv, !m.appOwned))...)
}

func NewCounterVec(appOwned bool, component, name, help string, ln []string) (*CounterVec, error) {
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

	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: NormalizeMetricName(component),
			Name:      NormalizeMetricName(name),
			Help:      help,
		},
		ln)

	if err = DefaultRegistrer().Register(vec); err != nil {
		return nil, err
	}

	return &CounterVec{vec: vec, appOwned: appOwned}, nil
}

func MustNewCounterVec(appOwned bool, component, name, help string, labelNames []string) *CounterVec {
	vec, err := NewCounterVec(appOwned, component, name, help, labelNames)
	if err != nil {
		panic(fmt.Errorf("unable to create new instance of Counter metric (%T); err: %w", &CounterVec{}, err))
	}

	return vec
}
