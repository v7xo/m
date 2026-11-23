package osutil_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

func Exec(t *testing.T, ctx context.Context, p osutil.ExecParams) {
	t.Helper()

	err := osutil.Exec(ctx, p)
	require.NoError(t, err, "unexpected error while executing command '%s'", stringsext.Stringer(p))
}
