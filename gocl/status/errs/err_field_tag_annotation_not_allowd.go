package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrFieldTagAnnotationNotAllowed = clerr.Errf("field shouldn't be annotated with such tag key")
