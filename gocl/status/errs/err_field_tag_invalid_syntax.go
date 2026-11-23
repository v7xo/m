package errs

import (
	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrFieldTagInvalidSyntax = clerr.Errf("%w; field tag value not conforms required syntax", ErrInvalidSyntax)
