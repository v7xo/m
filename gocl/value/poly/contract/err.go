package contract

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var (
	ErrContract     = clerr.ErrStatusf(opstatus.InternalFail, "poly-contract error")
	ErrRegistration = errs.Join(ErrContract, errors.New("registration fault"))
)
