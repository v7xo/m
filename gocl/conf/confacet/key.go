package confacet

import "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"

type Key string

const (
	KeyConfVariant                 = Key(attrname.ConfVariant)
	KeyDiagnosticsEndpointDisabled = Key(attrname.DiagnosticsEndpointDisabled)
	KeyLogAttrInclusion            = Key(attrname.LogAttrInclusion)
	KeyLogFormat                   = Key(attrname.LogFormat)
	KeyLogLevel                    = Key(attrname.LogLevel)
	KeyLogLibProvider              = Key(attrname.LogLibProvider)
)

func (v Key) String() string { return string(v) }
