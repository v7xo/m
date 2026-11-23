package level

import (
	"strings"
)

type Name string

const (
	ErrorNum  Num  = 0
	Error     Name = "ERR"
	ErrorLong Name = "ERROR"

	WarnNum   Num  = 1
	Warn      Name = "WRN"
	WarnLong1 Name = "WARNING"
	WarnLong2 Name = "WARN"

	InfoNum  Num  = 2
	Info     Name = "INF"
	InfoLong Name = "INFO"

	TraceOrDebugNum Num = 3

	DebugNum  Num  = TraceOrDebugNum
	Debug     Name = "DBG"
	DebugLong Name = "DEBUG"

	TraceNum  Num  = TraceOrDebugNum
	Trace     Name = "TRC"
	TraceLong Name = "TRACE"

	Unspecified Name = ""
)

func (v Name) IsTrace() bool {
	return strings.EqualFold(string(Trace), string(v)) || strings.EqualFold(string(TraceLong), string(v))
}

func (v Name) IsDebug() bool {
	return strings.EqualFold(string(Debug), string(v)) || strings.EqualFold(string(DebugLong), string(v))
}

func (v Name) IsTraceOrDebug() bool {
	return v.IsTrace() || v.IsDebug()
}

func (v Name) IsWarn() bool {
	return strings.EqualFold(string(Warn), string(v)) ||
		strings.EqualFold(string(WarnLong1), string(v)) ||
		strings.EqualFold(string(WarnLong2), string(v))
}

func (v Name) IsInfo() bool {
	return strings.EqualFold(string(Info), string(v)) || strings.EqualFold(string(InfoLong), string(v))
}

func (v Name) IsError() bool {
	return strings.EqualFold(string(Error), string(v)) ||
		strings.EqualFold(string(ErrorLong), string(v))
}

// Num returns numeric representation of this level v.
// In case of unknown name of level v [TraceNum] returned as a fallback.
func (v Name) Num() Num {
	switch {
	case v.IsTraceOrDebug():
		return TraceNum
	case v.IsInfo():
		return InfoNum
	case v.IsWarn():
		return WarnNum
	case v.IsError():
		return ErrorNum
	default:
		return TraceNum
	}
}

func (v Name) IsKnown() bool {
	return v.IsTraceOrDebug() || v.IsError() || v.IsWarn() || v.IsInfo()
}

func (v Name) IsUnspecified() bool {
	return v == Unspecified
}

func (v Name) String() string {
	return string(v)
}
