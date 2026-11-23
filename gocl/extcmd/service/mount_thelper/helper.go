package mount_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func New(t *testing.T, ctx context.Context, opts ...mount.OptionFn) mount.Mount {
	t.Helper()

	res, err := mount.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}

func Mount(t *testing.T, ctx context.Context, s mount.Mount, p mount.MountParams) (target string) {
	t.Helper()

	target, unmount, err := s.Mount(ctx, p)
	require.NoError(t, err, "unexpected error while mounting, params '%s'", stringsext.Stringer(p))
	require.NotNil(t, unmount, "unexpected nil unmount function reterned as a result of mount operation using params '%s'", stringsext.Stringer(p))
	require.NotEmpty(t, target, "unexpected empty auto target returned as a result of mount operation using params '%s'", stringsext.Stringer(p))

	clhlp.Cleanup(t, func() error { return unmount(ctx) }, "error while unmounting '%s'", target)

	return
}
