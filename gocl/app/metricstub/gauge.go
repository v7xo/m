package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

func NewGaugeVec(component, name, help string, labels []string) (*prommetric.GaugeVec, error) {
	return prommetric.NewGaugeVec(true, component, name, help, labels)
}

func MustNewGaugeVec(component, name, help string, labels []string) *prommetric.GaugeVec {
	return prommetric.MustNewGaugeVec(true, component, name, help, labels)
}
