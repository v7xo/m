package syncext

import "sync/atomic"

var _ = func() {
	var (
		s1 *atomic.Bool
		s2 *atomic.Int32
		s3 *atomic.Int64
		s4 *atomic.Uint32
		s5 *atomic.Uint64
	)

	// nolint:errcheck
	{
		CasRestoreOnErr(s1, true, true, nil, nil)
		CasRestoreOnErr(s2, 0, 0, nil, nil)
		CasRestoreOnErr(s3, 0, 0, nil, nil)
		CasRestoreOnErr(s4, 0, 0, nil, nil)
		CasRestoreOnErr(s5, 0, 0, nil, nil)
	}
}

func CasRestoreOnErr[
	L interface {
		*atomic.Bool | *atomic.Int32 | *atomic.Int64 | *atomic.Uint32 | *atomic.Uint64

		CompareAndSwap(o, n V) bool
		Store(v V)
	},
	V any](l L, o, n V, casErr error, fn func() error) error {
	if !l.CompareAndSwap(o, n) {
		return casErr
	}

	if fnErr := fn(); fnErr != nil {
		l.Store(o)

		return fnErr
	}

	return nil
}
