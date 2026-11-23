package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

func NewSummaryVec(component, name, help string, labels []string) (*prommetric.SummaryVec, error) {
	return prommetric.NewSummaryVec(true, component, name, help, labels)
}

func MustNewSummaryVec(component, name, help string, labels []string) *prommetric.SummaryVec {
	return prommetric.MustNewSummaryVec(true, component, name, help, labels)
}
