package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrDataContractInvalidVersion = clerr.ErrStatusf(opst.DataContractInvalidVersion, "data contract version not matches to expected")
