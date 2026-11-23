package model

import (
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type DataSizeUnit int32

const (
	DataSizeUnitUnspecified DataSizeUnit = iota
	DataSizeUnitB
	DataSizeUnitKB
	DataSizeUnitMB
)

const DataAmountInvalid int64 = -1

func (v DataSizeUnit) String() string {
	switch v {
	case DataSizeUnitUnspecified:
		return placeholder.Unspecified
	case DataSizeUnitB:
		return "B"
	case DataSizeUnitKB:
		return "KB"
	case DataSizeUnitMB:
		return "MB"
	default:
		return placeholder.UnknownEnumInt(v)
	}
}

type DataSize struct {
	Unit       DataSizeUnit
	UnitAmount int64
}

// Amount returns total amount in bytes.
func (v DataSize) Amount() int64 {
	if v.UnitAmount < 0 {
		return DataAmountInvalid
	}

	switch v.Unit {
	case DataSizeUnitB:
		return v.UnitAmount
	case DataSizeUnitKB:
		return v.UnitAmount * 1000
	case DataSizeUnitMB:
		return v.UnitAmount * 1000000
	case DataSizeUnitUnspecified:
		fallthrough
	default:
		return DataAmountInvalid
	}
}

func (v DataSize) String() string {
	return stringsext.FormatInt(v.UnitAmount) + v.Unit.String()
}
