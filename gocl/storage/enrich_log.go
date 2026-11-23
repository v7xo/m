package storage

import (
	"net"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

func EnrichLogWithConnConfProps(e logext.Event, conf ConnConf) logext.Event {
	return e.
		Str(attrname.Addr, net.JoinHostPort(conf.Host, strconv.Itoa(conf.Port))).
		Str(attrname.AddrHost, conf.Host).
		Int(attrname.AddrPort, conf.Port).
		Str(attrname.StorageName, conf.Database).
		Str(attrname.StorageUser, conf.User).
		Bool(attrname.StoragePooled, conf.PoolEnabled).
		Str(attrname.StorageType, conf.Type)
}
