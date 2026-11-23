package factorystub

import (
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	factorypkg "rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
)

type (
	Factory[TAPI storageapi.SQLAPI] = factorypkg.Factory[TAPI]
	ClientConf                      = storageapi.ClientConf
)
