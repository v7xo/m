package ratelimiter

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

const (
	logScope           = "rate_limiter"
	missingLastEventTs = int64(-1)

	MinJitter = float64(.1)

	DefaultJitter      = timeout.DefaultJitter
	DefaultName        = placeholder.Unassigned
	DefaultMinInterval = time.Second
)
