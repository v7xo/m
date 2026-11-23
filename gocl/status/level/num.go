package level

import (
	"log/syslog"
)

type Num byte

func (l Num) IsDebug() bool        { return l == DebugNum }
func (l Num) IsError() bool        { return l == ErrorNum }
func (l Num) IsInfo() bool         { return l == InfoNum }
func (l Num) IsTrace() bool        { return l == TraceNum }
func (l Num) IsTraceOrDebug() bool { return l == TraceOrDebugNum }
func (l Num) IsWarn() bool         { return l == WarnNum }

func (l Num) Name() Name {
	switch l {
	case DebugNum:
		return Debug
	case ErrorNum:
		return Error
	case InfoNum:
		return Info
	case WarnNum:
		return Warn
	default:
		return Trace
	}
}

func (l Num) IsKnown() bool {
	return l == TraceOrDebugNum || l == ErrorNum || l == WarnNum || l == InfoNum
}

func (l Num) String() string { return string(l.Name()) }

func FromSyslogLevel(l syslog.Priority) Num {
	//nolint:exhaustive
	switch l {
	case syslog.LOG_DEBUG:
		return DebugNum
	case syslog.LOG_NOTICE:
		fallthrough
	case syslog.LOG_INFO:
		return InfoNum
	case syslog.LOG_WARNING:
		return WarnNum
	case syslog.LOG_ALERT:
		fallthrough
	case syslog.LOG_CRIT:
		fallthrough
	case syslog.LOG_EMERG:
		fallthrough
	case syslog.LOG_ERR:
		return ErrorNum
	default:
		return InfoNum
	}
}
