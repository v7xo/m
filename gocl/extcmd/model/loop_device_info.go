package model

import "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"

type LoopDeviceInfo struct {
	BackingFile       string
	Name              string
	SizeLimit         int64
	Offset            int64
	LogicalSectorSize int16
	AutoClear         bool
	ReadOnly          bool
	DirectIO          bool
}

func (v LoopDeviceInfo) ToBlockDeviceID() BlockDeviceID {
	var backing BlockDeviceID

	if len(v.BackingFile) != 0 {
		backing = NewBlockDeviceID(v.BackingFile, BlockDeviceTypeCodeFile, nil)
	}

	return NewBlockDeviceID(v.Name, BlockDeviceTypeCodeLoop, BlockDeviceIDOption().Backing(backing))
}

func (v LoopDeviceInfo) String() string {
	return "(name: " + placeholder.UnspecifiedIfEmpty(v.Name) + ", backing: " + placeholder.UnspecifiedIfEmpty(v.BackingFile) + ")"
}

type LoopDeviceInfoList []LoopDeviceInfo

func (v LoopDeviceInfoList) FindByName(val string) (LoopDeviceInfo, int, bool) {
	for i := range v {
		if v[i].Name == val {
			return v[i], i, true
		}
	}

	return LoopDeviceInfo{}, -1, false
}
