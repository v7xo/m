package mount

import (
	"errors"
	"fmt"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	clstrings "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type UmountParams interface {
	isUnmountParams()

	String() string
}

var _ UmountParams = UnmountParamsTargetBase{}

type UnmountParamsTargetBase struct {
	Target string
}

func (UnmountParamsTargetBase) isUnmountParams() {}

func (v UnmountParamsTargetBase) String() string {
	return "(target: " + v.Target + ")"
}

func (v UnmountParamsTargetBase) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; unmount parameters invalid; err: %w", errs.ErrBadArgs, err)
		}
	}()

	if len(v.Target) == 0 {
		return errors.New("no target to unmount specified")
	}

	return nil
}

type MountParams interface {
	IsMountParams()

	String() string
}

var _ MountParams = MountParamsDeviceBase{}

type MountParamsDeviceBase struct {
	Source              model.BlockDeviceID
	Target              string
	Options             model.MountOptions
	Type                model.FileSystemTypeCode
	TakeTargetOwnership bool
}

func (v MountParamsDeviceBase) IsMountParams() {}

func (v MountParamsDeviceBase) String() string {
	sb := strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("type: ")
	sb.WriteString(v.Type.String())

	sb.WriteString(", target: ")
	sb.WriteString(v.Target)

	sb.WriteString(", source: ")
	sb.WriteString(clstrings.Stringer(v.Source))

	if opts := v.Options.String(); opts != "" {
		sb.WriteString(", options: ")
		sb.WriteString(v.Options.String())
	}

	sb.WriteByte(')')

	return sb.String()
}

func (v MountParamsDeviceBase) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; invalid mount parameters (type '%T'); err: %w", errs.ErrBadArgs, v, err)
		}
	}()

	if v.Source == nil {
		return errors.New("no source specified")
	}

	if _, ok := v.Type.ToLinuxKernelKnownName(); !ok && !v.Type.IsOther() && !v.Type.IsUnspecified() {
		return fmt.Errorf("file system type code '%s' can't be mapped to Linux kernel known file system type name", v.Type)
	}

	return
}
