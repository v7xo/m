package logext

import (
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

const (
	ConfFlenvSelfFlagPrefix = "log"

	AttrInclExclPatternREDesignator = flenv.RegexpDesignator

	DefaultLibProvider                    = LibProviderGoKit
	DefaultRequestBodyFieldLimitBytes     = 16384 - placeholder.OverlimitLen
	DefaultRequestHeadersFieldLimitBytes  = 8192 - placeholder.OverlimitLen
	DefaultResponseBodyFieldLimitBytes    = 16384 - placeholder.OverlimitLen
	DefaultResponseHeadersFieldLimitBytes = 8192 - placeholder.OverlimitLen
	DefaultTimestampFieldFormat           = "2006-01-02T15:04:05.999Z07:00"
)
