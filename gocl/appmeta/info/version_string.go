package info

import (
	"rn.m.vo7.io/vo7io/m/gocl/appmeta/buildinfo"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type VersionString string

func (v VersionString) IsDevel() bool {
	return v == buildinfo.DevelVersion
}

func (v VersionString) IsUnknown() bool {
	return v == "" || v == placeholder.Unknown
}

func (v VersionString) String() string { return string(v) }
