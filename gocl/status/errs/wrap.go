package errs

import (
	"fmt"
)

func Wrap(err, with error) error {
	if with == nil {
		return err
	} else if err == nil {
		return with
	}

	//nolint:revive
	return fmt.Errorf("%s; %w", with.Error(), err)
}
