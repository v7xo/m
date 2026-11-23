package subvol_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/subvol"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

func New(t *testing.T, ctx context.Context, opts ...subvol.OptionFn) subvol.Subvol {
	t.Helper()

	res, err := subvol.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}

func Create(t *testing.T, ctx context.Context, s subvol.Subvol, path string) {
	t.Helper()

	err := s.Create(ctx, path)
	require.NoError(t, err, "unexpected error while creating BTRFS subvolume '%s'", path)
}
