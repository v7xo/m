package prompusher

import (
	"time"
)

func DefaultConf() Conf {
	//nolint:mnd,revive
	return Conf{
		GatewayAddress: "http://localhost:9001",
		Interval:       time.Second * 5,
		Timeout:        time.Second * 3,
	}
}
