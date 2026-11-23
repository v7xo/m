package confset

import (
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type Set interface {
	isConfSet()

	Confs() ConfGetter

	flenv.Conf
	flenv.FlagSetBinderPrefixDefault
	flenv.UsePreparer
	logext.ConfGetter
	endpoint.ConfGetter
}
