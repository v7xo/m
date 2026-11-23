package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrDataContractViolation = clerr.ErrStatusf(opst.DataContractViolation, "got data not conform the required data contract")
