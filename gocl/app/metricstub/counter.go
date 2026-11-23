package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

func NewCounterVec(component, name, help string, labels []string) (*prommetric.CounterVec, error) {
	return prommetric.NewCounterVec(true, component, name, help, labels)
}

func MustNewCounterVec(component, name, help string, labels []string) *prommetric.CounterVec {
	return prommetric.MustNewCounterVec(true, component, name, help, labels)
}
