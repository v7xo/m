package losetup

import (
	"fmt"
)

type AttachParams interface {
	isAttachParams()

	String() string
}

type AttachParamsFile struct {
	BackingFile string
}

var _ AttachParams = AttachParamsFile{}

func (AttachParamsFile) isAttachParams() {}

func (v AttachParamsFile) String() string {
	return fmt.Sprintf("(backing file: %s)", v.BackingFile)
}
