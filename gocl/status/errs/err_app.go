package errs

import (
	"fmt"
	"os"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var (
	ErrAppAlreadyStarted     = clerr.Errf("already started")
	ErrAppCommandFailure     = clerr.Errf("command failed")
	ErrAppSubCommandRequired = clerr.Errf("subcommand required (see --help)")
)

func AppCommandFailureJoin(err error, includeArgs bool) error {
	if includeArgs {
		return Join(ErrAppCommandFailure, fmt.Errorf("command line args: %q", os.Args), err)
	}

	return Join(ErrAppCommandFailure, err)
}
