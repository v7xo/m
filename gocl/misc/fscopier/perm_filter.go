package fscopier

import (
	"io/fs"
	"os"
)

type (
	PermFilter interface {
		Calc(src, dst string) (os.FileMode, error)
	}

	PermFilterFn func(src, dst string) (os.FileMode, error)
)

var (
	_              PermFilter  = PermFilterFn(nil)
	file_GU_R_Perm fs.FileMode = 0o440
	file_GU_W_Perm fs.FileMode = 0o220
	file_GU_X_Perm fs.FileMode = 0o110
)

func (v PermFilterFn) Calc(src, dst string) (os.FileMode, error) { return v(src, dst) }

func (left PermFilterFn) Read() PermFilterFn {
	return func(src, dst string) (os.FileMode, error) {
		l, err := left(src, dst)
		if err != nil {
			return 0, err
		}

		return l | file_GU_R_Perm, nil
	}
}

func (left PermFilterFn) Write() PermFilterFn {
	return func(src, dst string) (os.FileMode, error) {
		l, err := left(src, dst)
		if err != nil {
			return 0, err
		}

		return l | file_GU_W_Perm, nil
	}
}

func (left PermFilterFn) Execute() PermFilterFn {
	return func(src, dst string) (os.FileMode, error) {
		l, err := left(src, dst)
		if err != nil {
			return 0, err
		}

		return l | file_GU_X_Perm, nil
	}
}

func (left PermFilterFn) Shell() PermFilterFn {
	return func(src, dst string) (os.FileMode, error) {
		l, err := left(src, dst)
		if err != nil {
			return 0, err
		}

		if shellFilePathRegexp.MatchString(src) {
			return l | file_GU_X_Perm, nil // add Executable bits
		}

		return l, nil
	}
}

func (left PermFilterFn) Add(right PermFilter) PermFilterFn {
	return func(src, dst string) (os.FileMode, error) {
		l, err := left(src, dst)
		if err != nil {
			return 0, err
		}

		r, err := right.Calc(src, dst)
		if err != nil {
			return 0, err
		}

		return l | r, nil
	}
}

func ZeroPermFilter() PermFilterFn {
	return func(_, _ string) (os.FileMode, error) { return 0, nil }
}
