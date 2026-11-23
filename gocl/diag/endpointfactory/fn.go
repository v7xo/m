package endpointfactory

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type Fn func(endpoint.ConfGetter, logext.Logger) (endpoint.Service, error)
