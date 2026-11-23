package osutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func Pushd(wd string) (popd func() error, err error) {
	prevDir, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to get current working directory; err: %w", err)

		return
	}

	switch {
	case wd == "" || wd == ".":
		wd = prevDir
	case !filepath.IsAbs(wd):
		if wd, err = filepath.Abs(wd); err != nil {
			err = fmt.Errorf("failed to make directory path '%s' absolute; err: %w", wd, err)

			return
		}
	}

	if err = os.Chdir(wd); err != nil {
		err = fmt.Errorf("failed to change working directory to '%s'; err: %w", wd, err)

		return
	}

	popd = func() error { return os.Chdir(prevDir) }

	return
}
