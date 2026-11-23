package prommetric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
)

var defaultBuckets = []float64{
	.003,
	.005,
	.01,
	.017,
	.025,
	.037,
	.05,
	.063,
	.075,
	.1,
	.137,
	.175,
	.25,
	.375,
	.5,
	.75,
	.875,
	1,
	1.25,
	1.5,
	2.0,
	2.5,
	3.75,
	5,
	10,
	15,
	20,
	25,
	30,
	60,
	90,
	120,
	180,
	300,
	600,
	900,
}

type HistogramVec struct {
	vec      *prometheus.HistogramVec
	appOwned bool
}

func (m *HistogramVec) WithLabelVals(lv ...string) prometheus.Observer {
	return m.vec.WithLabelValues(normalizeLabelVal(prependPredefinedLabelVal(lv, !m.appOwned))...)
}

func NewHistogramVec(appOwned bool, component, name, help string, ln []string, buckets ...float64) (*HistogramVec, error) {
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

	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: NormalizeMetricName(component),
			Name:      NormalizeMetricName(name),
			Help:      help,
			Buckets:   buckets,
		},
		ln)

	if err = DefaultRegistrer().Register(vec); err != nil {
		return nil, err
	}

	return &HistogramVec{vec: vec, appOwned: appOwned}, nil
}

func MustNewHistogramVec(appOwned bool, component, name, help string, labels []string, buckets ...float64) *HistogramVec {
	vec, err := NewHistogramVec(appOwned, component, name, help, labels, buckets...)
	if err != nil {
		panic(fmt.Errorf("unable to create new instance of Histogram metric (%T); err: %w", &HistogramVec{}, err))
	}

	return vec
}
