package backoff

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

const (
	ConfigFlagPrefix = "backoff"

	DefaultExpBackoffBaseDelay  = time.Second
	DefaultExpBackoffJitter     = timeout.DefaultJitter
	DefaultExpBackoffMaxDelay   = time.Second * 60
	DefaultExpBackoffMultiplier = 1.6
)
