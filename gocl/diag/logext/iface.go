package logext

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext/logtypes"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type ConfGetter = interface{ LogConf() (*Conf, error) }

type (
	Event     = logtypes.Event
	EventList []Event
)

func (v EventList) HasLevel(l level.Num) bool {
	for i := range v {
		if v[i].Level() == l {
			return true
		}
	}

	return false
}

func (v EventList) LevelCount(l level.Num) int64 {
	c := int64(0)

	for i := range v {
		if v[i].Level() == l {
			c++
		}
	}

	return c
}

type Logger = logtypes.Logger
