package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrFieldTagAmbiguous = clerr.Errf("field tag is ambiguous (same tag occurs twice at least)")
