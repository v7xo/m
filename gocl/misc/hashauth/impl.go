package hashauth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	salt []byte
	hash []byte
}

func (s *impl) Verify(ctx context.Context, secret string) (err error) {
	actualHash := sha256.Sum256(append(s.salt, []byte(secret)...))

	if slices.Equal(s.hash, actualHash[:]) {
		return nil
	}

	return fmt.Errorf("hash authentication failed; err: %w", errs.ErrUnauthenticated)
}

func New(opts ...OptionFn) (_ Verifier, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	return impl, nil
}
