package confstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type (
	Log          = logext.Conf
	DiagEndpoint = endpoint.Conf
	Start        = startconf.Conf
)
