package storage

import (
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type IsolationLevel byte

const (
	IsolationLevelUnspecified IsolationLevel = iota
	IsolationLevelReadUncommitted
	IsolationLevelReadCommitted
	IsolationLevelRepeatableRead
	IsolationLevelSerializable
	IsolationLevelInheritedOrDefault
)

func (v IsolationLevel) IsUnspecified() bool {
	return v == IsolationLevelUnspecified
}

func (v IsolationLevel) String() string {
	switch v {
	case IsolationLevelUnspecified:
		return placeholder.Unspecified
	case IsolationLevelReadUncommitted:
		return "read_uncommitted"
	case IsolationLevelReadCommitted:
		return "read_committed"
	case IsolationLevelRepeatableRead:
		return "repeatable_read"
	case IsolationLevelSerializable:
		return "serializable"
	case IsolationLevelInheritedOrDefault:
		return "inherited_or_default"
	default:
		return placeholder.UnknownEnumUint(v)
	}
}
