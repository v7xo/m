package prommetric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
)

type SummaryVec struct {
	vec      *prometheus.SummaryVec
	appOwned bool
}

func (m *SummaryVec) WithLabelVals(lv ...string) prometheus.Observer {
	return m.vec.WithLabelValues(normalizeLabelVal(prependPredefinedLabelVal(lv, !m.appOwned))...)
}

func NewSummaryVec(appOwned bool, component, name, help string, ln []string) (*SummaryVec, error) {
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

	vec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Subsystem: NormalizeMetricName(component),
			Name:      NormalizeMetricName(name),
			Help:      help,
		},
		ln)

	if err = DefaultRegistrer().Register(vec); err != nil {
		return nil, err
	}

	return &SummaryVec{vec: vec, appOwned: appOwned}, nil
}

func MustNewSummaryVec(appOwned bool, component, name, help string, labelNames []string) *SummaryVec {
	vec, err := NewSummaryVec(appOwned, component, name, help, labelNames)
	if err != nil {
		panic(fmt.Errorf("unable to create new instance of Summary metric (%T); err: %w", &SummaryVec{}, err))
	}

	return vec
}
