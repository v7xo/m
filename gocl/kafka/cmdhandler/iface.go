package cmdhandler

import (
	"rn.m.vo7.io/vo7io/m/gocl/app/appmodule"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
)

type Service interface {
	appmodule.Module
	readiness.Collector
}
