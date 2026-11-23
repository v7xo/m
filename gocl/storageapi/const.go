package storageapi

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

const (
	ConfFlenvSelfFlagPrefix = "storage-api"

	DefaultCommandTimeout              = time.Minute * 3
	DefaultMetricsDisabled             = false
	DefaultName                        = placeholder.Default
	DefaultQueryNamedParameterDisabled = false
)
