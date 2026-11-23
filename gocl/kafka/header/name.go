package header

import (
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

type Code int64

const (
	CodeCorrelationID Code = 1 << iota
	CodeTraceParent
	CodeContentType
	CodeSchemaVersion

	CodeUnspecified Code = 0
)

var _ = enum.MustRegisterFlag(
	map[Code]string{
		CodeCorrelationID: string(NameCorrelationID),
		CodeTraceParent:   string(NameTraceParent),
		CodeContentType:   string(NameContentType),
		CodeSchemaVersion: string(NameSchemaVersion),
		CodeUnspecified:   string(NameUnspecified),
	},
)

func (v Code) String() string { return enum.String(v) }

type Name string

const (
	NameUnspecified   = Name("")
	NameCorrelationID = Name("X-Correlation-Id")
	NameTraceParent   = Name("traceparent")
	NameContentType   = Name("Content-Type")
	NameSchemaVersion = Name("X-Schema-Version")
)

func (v Name) IsContentType() bool { return strings.EqualFold(string(v), string(NameContentType)) }

func (v Name) IsCorrelationID() bool { return strings.EqualFold(string(v), string(NameCorrelationID)) }

func (v Name) IsSchemaVersion() bool { return strings.EqualFold(string(v), string(NameSchemaVersion)) }

func (v Name) IsTraceParent() bool { return strings.EqualFold(string(v), string(NameTraceParent)) }

func (v Name) IsUnspecified() bool { return strings.EqualFold(string(v), string(NameUnspecified)) }

func (v Name) ToCode() Code {
	switch v {
	case NameCorrelationID:
		return CodeCorrelationID
	case NameTraceParent:
		return CodeTraceParent
	case NameContentType:
		return CodeContentType
	case NameSchemaVersion:
		return CodeSchemaVersion
	case NameUnspecified:
		fallthrough
	default:
		return CodeUnspecified
	}
}

func (v Name) ToAttrName() attrname.Name {
	switch v {
	case NameCorrelationID:
		return attrname.CorrelationID
	case NameTraceParent:
		return attrname.TraceParent
	case NameContentType:
		return attrname.ContentType
	case NameSchemaVersion:
		return attrname.SchemaVersion
	case NameUnspecified:
		fallthrough
	default:
		return attrname.Name(v)
	}
}

func (v Name) String() string { return string(v) }
