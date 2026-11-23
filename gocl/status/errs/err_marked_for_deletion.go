package errs

import (
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

var ErrMarkedForDeletion = clerr.ErrStatusf(opst.NotFound, "entity marked for deletion")

func MarkedForDeletionWrapWith(err error) error { return Wrap(ErrMarkedForDeletion, err) }
func IsMarkedForDeletion(err error) bool        { return errors.Is(err, ErrMarkedForDeletion) }
