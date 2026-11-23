package cmdstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

// SubCmdCode represents code of migration command: up, down, init, etc.
type SubCmdCode int

const (
	SubCmdCodeDown SubCmdCode = iota + 1
	SubCmdCodeInit
	SubCmdCodeStatus
	SubCmdCodeUp
)

func (v SubCmdCode) String() string {
	switch v {
	case SubCmdCodeInit:
		return initSubCmdName
	case SubCmdCodeStatus:
		return statusSubCmdName
	case SubCmdCodeUp:
		return upSubCmdName
	case SubCmdCodeDown:
		return downSubCmdName
	default:
		return placeholder.Unknown
	}
}
