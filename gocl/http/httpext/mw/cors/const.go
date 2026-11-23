package cors

import "time"

const (
	ConfigFlagPrefix = "cors"

	DefaultAllowCredentials   = false
	DefaultAllowedHeaders     = "*"
	DefaultAllowedMethods     = "HEAD,GET,POST"
	DefaultAllowedOrigins     = "*"
	DefaultMaxAge             = time.Second * 5
	DefaultOptionsPassthrough = false
)
