package cryptsetup

import (
	"errors"
	"fmt"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly"
)

type OpenParams struct {
	BackingDevice poly.Of[model.BlockDeviceID] `yaml:"backing_device,omitempty"`
	MapperName    string                       `yaml:"mapper_name,omitempty"`
	Options       stringsext.KVOpts            `yaml:"options,omitempty"`
}

func (v OpenParams) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.JoinIfNot(fmt.Errorf("not valid '%s'; %w", typestring.TypeStringForExpr(v), err), errs.ErrBadArgs)
		}
	}()

	if v.BackingDevice.Val == nil {
		return errors.New("no backing device specified")
	}

	return nil
}

func (v OpenParams) String() string {
	sb := &strings.Builder{}
	w := func(n, v string) {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString(n)
		sb.WriteString(": ")
		sb.WriteString(v)
	}

	sb.WriteByte('(')

	if !v.BackingDevice.IsZero() {
		w("backing device", stringsext.Stringer(v.BackingDevice.Val))
	}

	if v.MapperName != "" {
		w("mapper", v.MapperName)
	}

	if !v.Options.IsZero() {
		w("options", v.Options.String())
	}

	sb.WriteByte(')')

	return sb.String()
}
