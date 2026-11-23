package metricstub

import (
	clprom "rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

func NewHistogramVec(component, name, help string, labels []string, buckets ...float64) (*clprom.HistogramVec, error) {
	return clprom.NewHistogramVec(true, component, name, help, labels, buckets...)
}

func MustNewHistogramVec(component, name, help string, labels []string, buckets ...float64) *clprom.HistogramVec {
	return clprom.MustNewHistogramVec(true, component, name, help, labels, buckets...)
}
