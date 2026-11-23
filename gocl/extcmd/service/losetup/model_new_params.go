package losetup

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type NewParams interface {
	isNewParams()

	String() string
}

type NewParamsFile struct {
	BackingFile  string
	Size         model.DataSize
	RecreateFile bool
}

var _ NewParams = NewParamsFile{}

func (NewParamsFile) isNewParams() {}

func (v NewParamsFile) String() string {
	return fmt.Sprintf("(backing file: %s, size: %s, recreate backing file: %v)", v.BackingFile, v.Size, v.RecreateFile)
}
