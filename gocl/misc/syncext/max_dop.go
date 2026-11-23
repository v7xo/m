package syncext

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

const (
	MaxDOPAuto MaxDOP = iota - 1
	MaxDOPUnspecified
)

type MaxDOP int32

func (v MaxDOP) IsAuto() bool { return v == MaxDOPAuto }

func (v MaxDOP) IsUnspecified() bool { return v == MaxDOPUnspecified }

func (v MaxDOP) Validate() error {
	switch {
	case v == MaxDOPAuto:
		return nil
	case v == MaxDOPUnspecified:
		return fmt.Errorf("%w; value of max. degree of parallelism not specified", errs.ErrBadArgs)
	case v < MaxDOPUnspecified:
		return fmt.Errorf("%w; value '%d' of max. degree of parallelism is not valid, can't be less than 1", errs.ErrBadArgs, v)
	default:
		return nil
	}
}

func (v MaxDOP) String() string {
	if v.IsAuto() {
		return placeholder.Auto
	}

	if v.IsUnspecified() {
		return placeholder.Unspecified
	}

	return stringsext.FormatInt(v)
}

func (v MaxDOP) Decrement() MaxDOP {
	if v == MaxDOPAuto || v < MaxDOPUnspecified {
		return v
	}

	return v - 1
}

func (v MaxDOP) Effective() (int, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	if v == MaxDOPAuto {
		return runtime.GOMAXPROCS(-1), nil
	}

	return int(v), nil
}

func (v MaxDOP) MustEffective() int {
	res, err := v.Effective()
	if err != nil {
		panic(err)
	}

	return res
}

func MaxDOPRand() MaxDOP {
	n := rand.Int32() // nolint:gosec
	if n == 0 {
		return MaxDOPAuto
	}

	return MaxDOP(n)
}
