package subvol

import (
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type TransmitParams struct {
	SourceParent string
	Source       string
	Destination  string
}

func (v TransmitParams) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.JoinIfNot(fmt.Errorf("not valid '%s'; %w", typestring.TypeStringForExpr(v), err), errs.ErrBadArgs)
		}
	}()

	if v.Source == "" {
		return errors.New("no source subvolume specified")
	}

	return nil
}
