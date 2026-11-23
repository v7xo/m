package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrFieldTagValueAmbiguous = clerr.Errf("field tag value is ambiguous (tag with same value occurs twice at least)")
